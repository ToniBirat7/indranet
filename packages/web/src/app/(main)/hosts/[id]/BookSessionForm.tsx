'use client'

import { useState, FormEvent } from 'react'
import { useRouter } from 'next/navigation'
import { api } from '@/lib/api'
import { getToken } from '@/lib/auth'

interface Props {
  hostId: string
  pricePerHourCents: number
}

const DURATIONS = [
  { label: '15 min', minutes: 15 },
  { label: '30 min', minutes: 30 },
  { label: '1 hour', minutes: 60 },
  { label: '2 hours', minutes: 120 },
]

export default function BookSessionForm({ hostId, pricePerHourCents }: Props) {
  const router = useRouter()
  const [duration, setDuration] = useState(30)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Match the Go backend's integer division: ratePerMinuteCents = pricePerHourCents / 60
  const ratePerMinuteCents = Math.floor(pricePerHourCents / 60)
  const totalCents = ratePerMinuteCents * duration
  const totalDisplay = (totalCents / 100).toFixed(2)

  async function handleBook(e: FormEvent) {
    e.preventDefault()
    setError(null)

    const token = getToken()
    if (!token) {
      router.push(`/auth/login?return=/hosts/${hostId}`)
      return
    }

    setLoading(true)
    try {
      const res = await api.sessions.create({ host_id: hostId, duration_minutes: duration }, token)

      if (res.checkout_url) {
        // Production: redirect to Stripe Checkout
        window.location.href = res.checkout_url
      } else {
        // Dev mode (auto-authorized): go straight to session viewer
        router.push(`/session/${res.session_id}`)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create session')
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleBook} className="space-y-4">
      <div>
        <label className="block text-gray-300 text-sm mb-2">Session duration</label>
        <div className="grid grid-cols-2 gap-2">
          {DURATIONS.map((d) => (
            <button
              key={d.minutes}
              type="button"
              onClick={() => setDuration(d.minutes)}
              className={`py-2 rounded-lg text-sm font-medium transition-colors ${
                duration === d.minutes
                  ? 'bg-brand-600 text-white'
                  : 'bg-gray-800 text-gray-300 hover:bg-gray-700'
              }`}
            >
              {d.label}
            </button>
          ))}
        </div>
      </div>

      <div className="pt-2 border-t border-gray-800">
        <div className="flex justify-between text-sm text-gray-400 mb-1">
          <span>Session total</span>
          <span className="text-white font-medium">${totalDisplay}</span>
        </div>
      </div>

      {error && (
        <p className="text-red-400 text-sm">{error}</p>
      )}

      <button
        type="submit"
        disabled={loading}
        className="w-full bg-brand-600 hover:bg-brand-700 disabled:opacity-50 text-white font-semibold py-3 rounded-lg transition-colors"
      >
        {loading ? 'Creating session…' : 'Book Session'}
      </button>

      <p className="text-gray-600 text-xs text-center">
        You&apos;ll be charged only for actual session time
      </p>
    </form>
  )
}
