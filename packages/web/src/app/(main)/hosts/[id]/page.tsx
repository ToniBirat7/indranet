import { notFound } from 'next/navigation'
import BookSessionForm from './BookSessionForm'

interface Host {
  host_id: string
  display_name: string
  gpu_model: string
  vram_gb: number
  cpu_model: string
  ram_gb: number
  os: string
  price_per_hour_cents: number
  price_per_minute_cents: number
  online: boolean
  rating: number
  total_sessions: number
  tags: string[]
  payouts_enabled: boolean
}

interface Props {
  params: { id: string }
}

export default async function HostDetailPage({ params }: Props) {
  const apiURL = process.env.API_URL ?? process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080'

  const res = await fetch(`${apiURL}/v1/hosts/${params.id}`, { cache: 'no-store' })
  if (!res.ok) notFound()

  const host: Host = await res.json()
  const pricePerHour = (host.price_per_hour_cents / 100).toFixed(2)
  const pricePerMinute = (host.price_per_hour_cents / 6000).toFixed(3)

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Host specs */}
        <div className="lg:col-span-2">
          <div className="flex items-start justify-between mb-4">
            <div>
              <h1 className="text-3xl font-bold text-white">{host.display_name}</h1>
              {host.rating > 0 && (
                <span className="text-yellow-400 text-sm mt-1 inline-block">
                  ★ {host.rating.toFixed(1)} · {host.total_sessions} sessions
                </span>
              )}
            </div>
            <span className={`text-sm px-3 py-1 rounded-full ${
              host.online ? 'bg-green-900 text-green-400' : 'bg-gray-800 text-gray-500'
            }`}>
              {host.online ? 'Online' : 'Offline'}
            </span>
          </div>

          <div className="bg-gray-900 rounded-xl border border-gray-800 p-6 mb-6">
            <h2 className="text-white font-semibold mb-4">Hardware</h2>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <Spec label="GPU" value={host.gpu_model} />
              <Spec label="VRAM" value={`${host.vram_gb} GB`} />
              <Spec label="CPU" value={host.cpu_model} />
              <Spec label="RAM" value={`${host.ram_gb} GB`} />
              <Spec label="OS" value={host.os} />
            </div>
          </div>

          {host.tags.length > 0 && (
            <div className="flex flex-wrap gap-2 mb-6">
              {host.tags.map((tag) => (
                <span key={tag} className="bg-gray-800 text-gray-400 text-xs px-3 py-1 rounded-full">
                  {tag}
                </span>
              ))}
            </div>
          )}
        </div>

        {/* Booking panel */}
        <div className="lg:col-span-1">
          <div className="bg-gray-900 border border-gray-800 rounded-xl p-6 sticky top-24">
            <div className="mb-4">
              <span className="text-3xl font-bold text-white">${pricePerHour}</span>
              <span className="text-gray-400 text-sm">/hr</span>
              <p className="text-gray-500 text-xs mt-1">${pricePerMinute}/min</p>
            </div>

            {host.online ? (
              <BookSessionForm
                hostId={host.host_id}
                pricePerHourCents={host.price_per_hour_cents}
              />
            ) : (
              <div className="text-gray-500 text-sm text-center py-4">
                This host is currently offline.
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

function Spec({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-gray-500 text-xs mb-0.5">{label}</p>
      <p className="text-white">{value}</p>
    </div>
  )
}
