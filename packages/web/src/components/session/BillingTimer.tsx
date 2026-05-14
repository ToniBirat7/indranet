'use client'

import { useEffect, useState } from 'react'

interface Props {
  sessionId: string
}

// BillingTimer shows elapsed time and remaining balance in real-time.
// Updates every second on the client side (display only — actual billing is server-side).
export default function BillingTimer({ sessionId }: Props) {
  const [elapsed, setElapsed] = useState(0) // seconds
  const [minutesRemaining, setMinutesRemaining] = useState<number | null>(null)
  const [warning, setWarning] = useState(false)

  useEffect(() => {
    const timer = setInterval(() => setElapsed((s) => s + 1), 1000)
    return () => clearInterval(timer)
  }, [])

  // TODO: Subscribe to WebSocket session events for real remaining balance updates
  // ws.onmessage: handle session_warning event → setMinutesRemaining + setWarning(true)

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
      {minutesRemaining !== null && (
        <span className={warning ? 'text-yellow-400 font-semibold' : ''}>
          {warning && '⚠ '}
          {minutesRemaining} min remaining
        </span>
      )}
    </div>
  )
}
