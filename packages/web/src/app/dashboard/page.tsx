'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { api, UserProfile, SessionSummary } from '@/lib/api'
import { getToken } from '@/lib/auth'

const STATE_COLORS: Record<string, string> = {
  ACTIVE: 'text-green-400',
  AUTHORIZED: 'text-blue-400',
  CREATED: 'text-gray-400',
  ENDING: 'text-yellow-400',
  ENDED: 'text-gray-500',
  FAILED: 'text-red-400',
}

export default function DashboardPage() {
  const router = useRouter()
  const [profile, setProfile] = useState<UserProfile | null>(null)
  const [sessions, setSessions] = useState<SessionSummary[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const token = getToken()
    if (!token) {
      router.push('/auth/login?return=/dashboard')
      return
    }
    Promise.all([api.users.me(token), api.sessions.list(token)])
      .then(([p, s]) => {
        setProfile(p)
        setSessions(s.sessions)
      })
      .catch(() => router.push('/auth/login?return=/dashboard'))
      .finally(() => setLoading(false))
  }, [router])

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin w-8 h-8 border-2 border-brand-500 border-t-transparent rounded-full" />
      </div>
    )
  }

  const balanceDisplay = profile ? (profile.balance_cents / 100).toFixed(2) : '0.00'
  const hoursUsed = sessions
    .filter((s) => s.state === 'ENDED')
    .reduce((acc, s) => acc + s.total_charged_cents / (s.rate_per_minute_cents * 60 || 1), 0)
    .toFixed(1)

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-8">
        <h1 className="text-3xl font-bold text-white">Dashboard</h1>
        {profile?.host_id && (
          <Link href="/dashboard/host" className="text-brand-400 hover:text-brand-300 text-sm">
            Host dashboard →
          </Link>
        )}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="bg-gray-900 rounded-lg p-6">
          <p className="text-gray-400 text-sm">Wallet Balance</p>
          <p className="text-2xl font-bold text-white">${balanceDisplay}</p>
        </div>
        <div className="bg-gray-900 rounded-lg p-6">
          <p className="text-gray-400 text-sm">Total Sessions</p>
          <p className="text-2xl font-bold text-white">{sessions.length}</p>
        </div>
        <div className="bg-gray-900 rounded-lg p-6">
          <p className="text-gray-400 text-sm">Hours Used</p>
          <p className="text-2xl font-bold text-white">{hoursUsed}</p>
        </div>
      </div>

      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-semibold text-white">Recent Sessions</h2>
        <Link href="/" className="text-brand-400 hover:text-brand-300 text-sm">
          Browse hosts →
        </Link>
      </div>

      {sessions.length === 0 ? (
        <div className="bg-gray-900 rounded-lg p-8 text-center">
          <p className="text-gray-500 mb-3">No sessions yet.</p>
          <Link href="/" className="bg-brand-600 hover:bg-brand-700 text-white px-5 py-2 rounded-lg text-sm inline-block">
            Find a host
          </Link>
        </div>
      ) : (
        <div className="space-y-2">
          {sessions.map((s) => {
            const cost = (s.total_charged_cents / 100).toFixed(2)
            const date = new Date(s.created_at).toLocaleDateString()
            const isActive = s.state === 'ACTIVE' || s.state === 'AUTHORIZED'
            return (
              <div
                key={s.session_id}
                className="bg-gray-900 rounded-lg px-4 py-3 flex items-center justify-between"
              >
                <div>
                  <span className={`text-xs font-medium ${STATE_COLORS[s.state] ?? 'text-gray-400'}`}>
                    {s.state}
                  </span>
                  <p className="text-gray-400 text-xs mt-0.5">{date}</p>
                </div>
                <div className="text-right">
                  <p className="text-white text-sm">${cost}</p>
                  {isActive && (
                    <Link
                      href={`/session/${s.session_id}`}
                      className="text-brand-400 hover:text-brand-300 text-xs"
                    >
                      Rejoin →
                    </Link>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
