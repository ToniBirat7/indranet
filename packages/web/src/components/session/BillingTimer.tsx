'use client'

import { useEffect, useState } from 'react'

interface Props {
  sessionId: string
  minutesRemaining?: number | null
  warning?: boolean
  startedAt?: string | null
}

export default function BillingTimer({ sessionId: _sessionId, minutesRemaining, warning, startedAt }: Props) {
  const [elapsed, setElapsed] = useState(() => {
    if (startedAt) {
      const ms = Date.now() - new Date(startedAt).getTime()
      return Math.max(0, Math.floor(ms / 1000))
    }
    return 0
  })

  useEffect(() => {
    const timer = setInterval(() => setElapsed((s) => s + 1), 1000)
    return () => clearInterval(timer)
  }, [])

  const minutes = Math.floor(elapsed / 60)
  const seconds = elapsed % 60

  return (
    <div className="flex items-center gap-6 text-sm text-gray-400">
      <span>
        Elapsed:{' '}
        <span className="text-white font-mono">
          {String(minutes).padStart(2, '0')}:{String(seconds).padStart(2, '0')}
        </span>
      </span>
      {minutesRemaining !== null && minutesRemaining !== undefined && (
        <span className={warning ? 'text-yellow-400 font-semibold animate-pulse' : 'text-gray-400'}>
          {warning && '⚠ '}
          {minutesRemaining} min remaining
        </span>
      )}
    </div>
  )
}
