import { useState, FormEvent } from 'react'

interface Props {
  onConnect: (sessionId: string) => void
}

export default function UserMode({ onConnect }: Props) {
  const [sessionId, setSessionId] = useState('')

  function handleConnect(e: FormEvent) {
    e.preventDefault()
    if (sessionId.trim()) onConnect(sessionId.trim())
  }

  return (
    <div style={{ padding: '2rem', maxWidth: '42rem', margin: '0 auto' }}>
      <h1 style={{ fontSize: '1.5rem', fontWeight: 'bold', marginBottom: '1.5rem' }}>User Mode</h1>
      <p style={{ color: '#9ca3af', marginBottom: '1.5rem' }}>
        Browse the marketplace at{' '}
        <a href="https://indranet.io" style={{ color: '#60a5fa' }} target="_blank" rel="noreferrer">
          indranet.io
        </a>{' '}
        and paste your session ID below to connect.
      </p>

      <form onSubmit={handleConnect} style={{ display: 'flex', gap: '0.75rem' }}>
        <input
          type="text"
          placeholder="Session ID (e.g. ses_abc123...)"
          value={sessionId}
          onChange={(e) => setSessionId(e.target.value)}
          style={{
            flex: 1,
            background: '#1f2937',
            border: '1px solid #4b5563',
            borderRadius: '0.375rem',
            padding: '0.5rem 0.75rem',
            fontSize: '0.875rem',
            color: 'white',
          }}
        />
        <button
          type="submit"
          disabled={!sessionId.trim()}
          style={{
            background: '#2563eb',
            color: 'white',
            padding: '0.5rem 1.25rem',
            borderRadius: '0.375rem',
            border: 'none',
            cursor: sessionId.trim() ? 'pointer' : 'not-allowed',
            opacity: sessionId.trim() ? 1 : 0.5,
          }}
        >
          Connect
        </button>
      </form>
    </div>
  )
}
