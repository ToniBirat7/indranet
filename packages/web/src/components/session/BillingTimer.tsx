'use client'

import { useEffect, useState } from 'react'

interface Props {
  sessionId: string
  minutesRemaining?: number | null
  warning?: boolean
}

export default function BillingTimer({ sessionId: _sessionId, minutesRemaining, warning }: Props) {
  const [elapsed, setElapsed] = useState(0)

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
