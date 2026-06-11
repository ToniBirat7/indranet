'use client'

import { useEffect, useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
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
  const [session, setSession] = useState<SessionState | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [minutesRemaining, setMinutesRemaining] = useState<number | null>(null)
  const [warning, setWarning] = useState(false)
  const [ended, setEnded] = useState(false)

  // Poll session state until ACTIVE or terminal
  useEffect(() => {
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
  }, [params.id, router])

  const handleBillingEvent = useCallback((event: BillingEvent) => {
    if (event.type === 'session_warning' && event.minutes_remaining !== undefined) {
      setMinutesRemaining(event.minutes_remaining)
      setWarning(true)
    } else if (event.type === 'session_kill' || event.type === 'session_failed') {
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
    return (
      <div className="flex items-center justify-center h-screen bg-gray-950 text-white">
        <div className="text-center">
          <h2 className="text-2xl font-bold mb-2">Session ended</h2>
          <p className="text-gray-400 mb-6">Your session has finished.</p>
          <button
            onClick={() => router.push('/dashboard')}
            className="bg-brand-600 hover:bg-brand-700 px-6 py-2 rounded-lg"
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
