'use client'

import { useState } from 'react'
import { invoke } from '@tauri-apps/api/core'

interface SystemInfo {
  gpu_vendor: string
  gpu_model: string
  ram_gb: number
  cpu_cores: number
  os: string
}

export default function HostMode() {
  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null)
  const [agentStatus, setAgentStatus] = useState<string>('stopped')
  const [error, setError] = useState<string | null>(null)

  async function loadSystemInfo() {
    try {
      const info = await invoke<SystemInfo>('get_system_info')
      setSystemInfo(info)
    } catch (e) {
      setError(String(e))
    }
  }

  async function startHosting() {
    // TODO: Authenticate with backend, get session assignment, start agent
    try {
      await invoke('start_agent', {
        sessionId: 'test-session',
        backendUrl: import.meta.env.VITE_API_URL ?? 'http://localhost:8080',
        token: localStorage.getItem('token') ?? '',
      })
      setAgentStatus('running')
    } catch (e) {
      setError(String(e))
    }
  }

  async function stopHosting() {
    try {
      await invoke('stop_agent')
      setAgentStatus('stopped')
    } catch (e) {
      setError(String(e))
    }
  }

  return (
    <div className="p-8 max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">Host Mode</h1>

      {error && (
        <div className="bg-red-900/50 text-red-300 p-4 rounded mb-4">{error}</div>
      )}

      <div className="bg-gray-800 rounded-lg p-6 mb-6">
        <h2 className="text-lg font-semibold mb-4">System</h2>
        {systemInfo ? (
          <dl className="grid grid-cols-2 gap-2 text-sm">
            <dt className="text-gray-400">GPU</dt>
            <dd>{systemInfo.gpu_vendor} {systemInfo.gpu_model}</dd>
            <dt className="text-gray-400">RAM</dt>
            <dd>{systemInfo.ram_gb} GB</dd>
            <dt className="text-gray-400">CPU Cores</dt>
            <dd>{systemInfo.cpu_cores}</dd>
            <dt className="text-gray-400">OS</dt>
            <dd>{systemInfo.os}</dd>
          </dl>
        ) : (
          <button onClick={loadSystemInfo} className="text-blue-400 hover:underline text-sm">
            Load system info
          </button>
        )}
      </div>

      <div className="flex gap-4">
        {agentStatus === 'stopped' ? (
          <button
            onClick={startHosting}
            className="bg-green-600 hover:bg-green-700 text-white px-6 py-2 rounded"
          >
            Start Hosting
          </button>
        ) : (
          <button
            onClick={stopHosting}
            className="bg-red-600 hover:bg-red-700 text-white px-6 py-2 rounded"
          >
            Stop Hosting
          </button>
        )}
      </div>

      <p className="mt-4 text-sm text-gray-500">
        Agent status: <span className="font-mono">{agentStatus}</span>
      </p>
    </div>
  )
}
