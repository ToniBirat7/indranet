'use client'

import { useState } from 'react'

export default function UserMode() {
  const [sessionId, setSessionId] = useState('')

  return (
    <div className="p-8 max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">User Mode</h1>
      <p className="text-gray-400 mb-6">
        Browse the marketplace at{' '}
        <a href="#" className="text-blue-400 hover:underline">
          indranet.io
        </a>{' '}
        and paste your session ID below to connect.
      </p>

      <div className="flex gap-3">
        <input
          type="text"
          placeholder="Session ID"
          value={sessionId}
          onChange={(e) => setSessionId(e.target.value)}
          className="flex-1 bg-gray-800 border border-gray-600 rounded px-3 py-2 text-sm"
        />
        <button
          disabled={!sessionId}
          className="bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white px-5 py-2 rounded"
        >
          Connect
        </button>
      </div>
      {/* TODO: Navigate to /session/<id> view when connected */}
    </div>
  )
}
