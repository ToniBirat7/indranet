import { useState } from 'react'
import HostMode from './app/pages/HostMode'
import UserMode from './app/pages/UserMode'
import Session from './app/pages/Session'

type View =
  | { page: 'home' }
  | { page: 'host' }
  | { page: 'user' }
  | { page: 'session'; sessionId: string; signalingUrl: string; token: string }

const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
const WS_URL = API_URL.replace(/^http/, 'ws')

export default function App() {
  const [view, setView] = useState<View>({ page: 'home' })

  function navigateToSession(sessionId: string) {
    const token = localStorage.getItem('indranet_token') ?? ''
    setView({
      page: 'session',
      sessionId,
      signalingUrl: `${WS_URL}/v1/signal/${sessionId}`,
      token,
    })
  }

  if (view.page === 'host') return <HostMode />
  if (view.page === 'user') return <UserMode onConnect={navigateToSession} />
  if (view.page === 'session') {
    return (
      <Session
        sessionId={view.sessionId}
        signalingUrl={view.signalingUrl}
        token={view.token}
      />
    )
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100vh', gap: '1rem' }}>
      <h1 style={{ fontSize: '2rem', fontWeight: 'bold', color: '#a78bfa' }}>IndraNet</h1>
      <p style={{ color: '#9ca3af', marginBottom: '1.5rem' }}>Peer-to-peer GPU compute</p>
      <div style={{ display: 'flex', gap: '1rem' }}>
        <button
          onClick={() => setView({ page: 'host' })}
          style={{ background: '#6d28d9', color: 'white', padding: '0.75rem 2rem', borderRadius: '0.5rem', border: 'none', cursor: 'pointer', fontSize: '1rem' }}
        >
          Host Mode
        </button>
        <button
          onClick={() => setView({ page: 'user' })}
          style={{ background: '#374151', color: 'white', padding: '0.75rem 2rem', borderRadius: '0.5rem', border: 'none', cursor: 'pointer', fontSize: '1rem' }}
        >
          User Mode
        </button>
      </div>
    </div>
  )
}
