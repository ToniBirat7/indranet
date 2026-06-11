'use client'

import { useEffect, useState, FormEvent } from 'react'
import { useRouter } from 'next/navigation'
import { api, UserProfile } from '@/lib/api'
import { getToken } from '@/lib/auth'

export default function HostDashboardPage() {
  const router = useRouter()
  const [profile, setProfile] = useState<UserProfile | null>(null)
  const [loading, setLoading] = useState(true)
  const [registering, setRegistering] = useState(false)
  const [registerError, setRegisterError] = useState<string | null>(null)
  const [agentToken, setAgentToken] = useState<string | null>(null)

  // Form state
  const [displayName, setDisplayName] = useState('')
  const [gpuModel, setGpuModel] = useState('')
  const [vramGb, setVramGb] = useState(8)
  const [cpuModel, setCpuModel] = useState('')
  const [ramGb, setRamGb] = useState(32)
  const [os, setOs] = useState('Windows 11')
  const [pricePerHour, setPricePerHour] = useState('2.00')
  const [tags, setTags] = useState('')

  useEffect(() => {
    const token = getToken()
    if (!token) {
      router.push('/auth/login?return=/dashboard/host')
      return
    }
    api.users.me(token)
      .then(setProfile)
      .catch(() => router.push('/auth/login?return=/dashboard/host'))
      .finally(() => setLoading(false))
  }, [router])

  async function handleRegister(e: FormEvent) {
    e.preventDefault()
    setRegisterError(null)

    const token = getToken()
    if (!token) return

    const priceCents = Math.round(parseFloat(pricePerHour) * 100)
    if (isNaN(priceCents) || priceCents <= 0) {
      setRegisterError('Invalid price')
      return
    }

    setRegistering(true)
    try {
      const res = await api.hosts.register({
        display_name: displayName,
        gpu_model: gpuModel,
        vram_gb: vramGb,
        cpu_model: cpuModel,
        ram_gb: ramGb,
        os,
        price_per_hour_cents: priceCents,
        tags: tags.split(',').map((t) => t.trim()).filter(Boolean),
      }, token)
      setAgentToken(res.agent_token)
      // Refresh profile to get host_id
      const updated = await api.users.me(token)
      setProfile(updated)
    } catch (err) {
      setRegisterError(err instanceof Error ? err.message : 'Registration failed')
    } finally {
      setRegistering(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin w-8 h-8 border-2 border-brand-500 border-t-transparent rounded-full" />
      </div>
    )
  }

  return (
    <div className="max-w-3xl mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold text-white mb-2">Host Dashboard</h1>
      <p className="text-gray-400 mb-8">
        Register your machine to start earning by renting GPU compute time.
      </p>

      {profile?.host_id ? (
        <RegisteredView hostId={profile.host_id} agentToken={agentToken} />
      ) : (
        <form onSubmit={handleRegister} className="space-y-6">
          <div className="bg-gray-900 rounded-xl border border-gray-800 p-6 space-y-4">
            <h2 className="text-white font-semibold">Machine Details</h2>

            <Field label="Display name" required>
              <input
                required value={displayName} onChange={(e) => setDisplayName(e.target.value)}
                placeholder="Alice's Gaming Rig"
                className={inputClass}
              />
            </Field>

            <div className="grid grid-cols-2 gap-4">
              <Field label="GPU model" required>
                <input
                  required value={gpuModel} onChange={(e) => setGpuModel(e.target.value)}
                  placeholder="RTX 4090"
                  className={inputClass}
                />
              </Field>
              <Field label="VRAM (GB)">
                <input
                  type="number" min={1} value={vramGb}
                  onChange={(e) => setVramGb(parseInt(e.target.value))}
                  className={inputClass}
                />
              </Field>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <Field label="CPU model">
                <input
                  value={cpuModel} onChange={(e) => setCpuModel(e.target.value)}
                  placeholder="Intel i9-13900K"
                  className={inputClass}
                />
              </Field>
              <Field label="RAM (GB)">
                <input
                  type="number" min={1} value={ramGb}
                  onChange={(e) => setRamGb(parseInt(e.target.value))}
                  className={inputClass}
                />
              </Field>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <Field label="Operating System">
                <select
                  value={os} onChange={(e) => setOs(e.target.value)}
                  className={inputClass}
                >
                  <option>Windows 11</option>
                  <option>Windows 10</option>
                  <option>Ubuntu 22.04</option>
                </select>
              </Field>
              <Field label="Price per hour (USD)">
                <div className="relative">
                  <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">$</span>
                  <input
                    type="number" step="0.01" min="0.01" value={pricePerHour}
                    onChange={(e) => setPricePerHour(e.target.value)}
                    className={`${inputClass} pl-7`}
                  />
                </div>
              </Field>
            </div>

            <Field label="Tags" description="Comma-separated: gaming, 4K, AI, etc.">
              <input
                value={tags} onChange={(e) => setTags(e.target.value)}
                placeholder="gaming, 4K, Stable Diffusion"
                className={inputClass}
              />
            </Field>
          </div>

          {registerError && (
            <p className="text-red-400 text-sm">{registerError}</p>
          )}

          <button
            type="submit"
            disabled={registering}
            className="w-full bg-brand-600 hover:bg-brand-700 disabled:opacity-50 text-white font-semibold py-3 rounded-lg transition-colors"
          >
            {registering ? 'Registering…' : 'Register host machine'}
          </button>
        </form>
      )}
    </div>
  )
}

function RegisteredView({ hostId, agentToken }: { hostId: string; agentToken: string | null }) {
  return (
    <div className="space-y-6">
      <div className="bg-green-900/20 border border-green-700 rounded-xl p-6">
        <h2 className="text-green-400 font-semibold mb-1">Host registered</h2>
        <p className="text-gray-400 text-sm">Host ID: <span className="font-mono text-white">{hostId}</span></p>
      </div>

      {agentToken && (
        <div className="bg-gray-900 rounded-xl border border-gray-800 p-6">
          <h3 className="text-white font-semibold mb-2">Agent token</h3>
          <p className="text-gray-400 text-sm mb-3">
            Copy this token to the host agent config. It will not be shown again.
          </p>
          <textarea
            readOnly
            value={agentToken}
            rows={3}
            className="w-full bg-gray-800 text-green-400 font-mono text-xs p-3 rounded-lg border border-gray-700 resize-none"
            onClick={(e) => (e.target as HTMLTextAreaElement).select()}
          />
        </div>
      )}

      <div className="bg-gray-900 rounded-xl border border-gray-800 p-6">
        <h3 className="text-white font-semibold mb-3">Next steps</h3>
        <ol className="text-gray-400 text-sm space-y-2 list-decimal list-inside">
          <li>Download the IndraNet host agent for Windows</li>
          <li>Paste the agent token into the agent config</li>
          <li>Run the agent — it will register your GPU and go online</li>
          <li>Users can now discover and book your machine</li>
        </ol>
      </div>
    </div>
  )
}

const inputClass =
  'w-full bg-gray-800 border border-gray-700 focus:border-brand-500 text-white rounded-lg px-3 py-2 text-sm outline-none transition-colors'

function Field({
  label,
  description,
  required,
  children,
}: {
  label: string
  description?: string
  required?: boolean
  children: React.ReactNode
}) {
  return (
    <div>
      <label className="block text-gray-300 text-sm mb-1">
        {label} {required && <span className="text-red-500">*</span>}
      </label>
      {description && <p className="text-gray-600 text-xs mb-1">{description}</p>}
      {children}
    </div>
  )
}
