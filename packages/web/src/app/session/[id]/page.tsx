'use client'

import { useEffect, useState, useCallback } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import StreamViewer from '@/components/session/StreamViewer'
import BillingTimer from '@/components/session/BillingTimer'
import { getToken } from '@/lib/auth'
import { api } from '@/lib/api'

interface SessionState {
  session_id: string
  state: string
  rate_per_minute_cents: number
  balance_remaining_minutes: number
}

type BillingEvent = {
  type: string
  minutes_remaining?: number
  reason?: string
}

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080'
const WS_URL = API_URL.replace(/^http/, 'ws')

export default function SessionPage({ params }: { params: { id: string } }) {
  const router = useRouter()
  const searchParams = useSearchParams()
  const paymentStatus = searchParams.get('payment') // 'success' | 'cancelled' | null
  const [session, setSession] = useState<SessionState | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [minutesRemaining, setMinutesRemaining] = useState<number | null>(null)
  const [warning, setWarning] = useState(false)
  const [ended, setEnded] = useState(false)
  const [endReason, setEndReason] = useState<string | null>(null)
  const [paymentBanner, setPaymentBanner] = useState(paymentStatus === 'success')
  const [rating, setRating] = useState(0)
  const [ratingSubmitted, setRatingSubmitted] = useState(false)

  useEffect(() => {
    if (!paymentBanner) return
    const t = setTimeout(() => setPaymentBanner(false), 4000)
    return () => clearTimeout(t)
  }, [paymentBanner])

  // Redirect immediately if payment was cancelled
  useEffect(() => {
    if (paymentStatus === 'cancelled') {
      router.push('/')
    }
  }, [paymentStatus, router])

  // Poll session state until ACTIVE or terminal
  useEffect(() => {
    if (paymentStatus === 'cancelled') return

    const token = getToken()
    if (!token) {
      router.push(`/auth/login?return=/session/${params.id}`)
      return
    }

    let active = true
    async function poll() {
      try {
        const s = await api.sessions.get(params.id, token!) as SessionState
        if (!active) return
        setSession(s)

        if (s.state === 'ENDED' || s.state === 'FAILED') {
          setEnded(true)
          if (s.state === 'FAILED') setEndReason('Session failed to start or was terminated unexpectedly.')
          return
        }
        if (s.state !== 'ACTIVE') {
          setTimeout(poll, 2000) // keep polling until ACTIVE
        } else {
          setMinutesRemaining(s.balance_remaining_minutes)
        }
      } catch (err) {
        if (active) setLoadError(err instanceof Error ? err.message : 'Failed to load session')
      }
    }

    poll()
    return () => { active = false }
  }, [params.id, router, paymentStatus])

  const handleBillingEvent = useCallback((event: BillingEvent) => {
    if (event.type === 'session_warning' && event.minutes_remaining !== undefined) {
      setMinutesRemaining(event.minutes_remaining)
      setWarning(true)
    } else if (event.type === 'session_kill') {
      if (event.reason === 'balance_exhausted') setEndReason('Your balance ran out.')
      setEnded(true)
    } else if (event.type === 'session_failed') {
      const reasons: Record<string, string> = {
        agent_timeout: 'The host took too long to start the session.',
        payment_failed: 'Payment was not completed.',
      }
      setEndReason(reasons[event.reason ?? ''] ?? 'Session failed unexpectedly.')
      setEnded(true)
    }
  }, [])

  if (loadError) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-950 text-white">
        <div className="text-center">
          <p className="text-red-400 text-lg mb-4">{loadError}</p>
          <button onClick={() => router.push('/')} className="text-brand-400 hover:text-brand-300">
            Back to marketplace
          </button>
        </div>
      </div>
    )
  }

  if (ended) {
    const canRate = !endReason && !ratingSubmitted // only prompt rating on clean end
    return (
      <div className="flex items-center justify-center h-screen bg-gray-950 text-white">
        <div className="text-center max-w-sm">
          <h2 className="text-2xl font-bold mb-2">Session ended</h2>
          <p className="text-gray-400 mb-6">{endReason ?? 'Your session has finished.'}</p>
          {canRate && (
            <div className="mb-6">
              <p className="text-gray-300 text-sm mb-3">How was the experience?</p>
              <div className="flex justify-center gap-2 mb-3">
                {[1, 2, 3, 4, 5].map((s) => (
                  <button
                    key={s}
                    onClick={() => setRating(s)}
                    className={`text-2xl transition-colors ${s <= rating ? 'text-yellow-400' : 'text-gray-600 hover:text-yellow-500'}`}
                  >
                    ★
                  </button>
                ))}
              </div>
              {rating > 0 && (
                <button
                  onClick={async () => {
                    const token = getToken()
                    if (token) {
                      await api.sessions.rate(params.id, rating, token).catch(() => {})
                      setRatingSubmitted(true)
                    }
                  }}
                  className="bg-brand-600 hover:bg-brand-700 px-5 py-1.5 rounded-lg text-sm mb-2 transition-colors"
                >
                  Submit rating
                </button>
              )}
            </div>
          )}
          {ratingSubmitted && <p className="text-green-400 text-sm mb-4">Thanks for the feedback!</p>}
          <button
            onClick={() => router.push('/dashboard')}
            className="bg-gray-800 hover:bg-gray-700 px-6 py-2 rounded-lg text-sm transition-colors"
          >
            View dashboard
          </button>
        </div>
      </div>
    )
  }

  if (!session || session.state !== 'ACTIVE') {
    const stateLabel: Record<string, string> = {
      CREATED: 'Waiting for payment…',
      AUTHORIZED: 'Setting up sandbox…',
      ENDING: 'Session ending…',
    }
    return (
      <div className="flex items-center justify-center h-screen bg-gray-950 text-white">
        {paymentBanner && (
          <div className="fixed top-4 left-1/2 -translate-x-1/2 z-50 bg-green-600 text-white px-6 py-3 rounded-lg shadow-lg text-sm font-medium animate-fade-in">
            Payment confirmed — starting your session
          </div>
        )}
        <div className="text-center">
          <div className="animate-spin w-10 h-10 border-2 border-brand-500 border-t-transparent rounded-full mx-auto mb-4" />
          <p className="text-gray-300 text-lg">
            {session ? (stateLabel[session.state] ?? 'Connecting…') : 'Loading session…'}
          </p>
          <p className="text-gray-600 text-sm mt-2">This usually takes under 30 seconds</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-screen bg-black">
      <div className="flex-1 relative">
        <StreamViewer
          sessionId={params.id}
          signalingUrl={`${WS_URL}/v1/signal/${params.id}`}
          onSessionEvent={handleBillingEvent}
        />
      </div>
      <div className="h-12 bg-gray-900 border-t border-gray-800 flex items-center px-4 justify-between">
        <BillingTimer
          sessionId={params.id}
          minutesRemaining={minutesRemaining}
          warning={warning}
        />
        <button
          onClick={async () => {
            const token = getToken()
            if (token) await api.sessions.end(params.id, token).catch(() => {})
            setEnded(true)
          }}
          className="text-gray-500 hover:text-red-400 text-sm transition-colors"
        >
          End session
        </button>
      </div>
    </div>
  )
}
