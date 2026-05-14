import Link from 'next/link'

interface Host {
  host_id: string
  display_name: string
  gpu_model: string
  vram_gb: number
  price_per_hour_cents: number
  online: boolean
  rating: number
  tags: string[]
}

export default function HostCard({ host }: { host: Host }) {
  const pricePerHour = (host.price_per_hour_cents / 100).toFixed(2)

  return (
    <Link href={`/hosts/${host.host_id}`}>
      <div className="bg-gray-900 border border-gray-800 hover:border-brand-500 rounded-xl p-5 transition-all cursor-pointer">
        <div className="flex items-start justify-between mb-3">
          <h3 className="text-white font-semibold truncate mr-2">{host.display_name}</h3>
          <span className={`text-xs px-2 py-1 rounded-full flex-shrink-0 ${
            host.online ? 'bg-green-900 text-green-400' : 'bg-gray-800 text-gray-500'
          }`}>
            {host.online ? 'Online' : 'Offline'}
          </span>
        </div>

        <p className="text-brand-400 font-medium mb-1">{host.gpu_model}</p>
        <p className="text-gray-400 text-sm mb-3">{host.vram_gb}GB VRAM</p>

        <div className="flex flex-wrap gap-1 mb-4">
          {host.tags.slice(0, 3).map((tag) => (
            <span key={tag} className="text-xs bg-gray-800 text-gray-400 px-2 py-0.5 rounded">
              {tag}
            </span>
          ))}
        </div>

        <div className="flex items-center justify-between">
          <span className="text-white font-bold text-lg">${pricePerHour}<span className="text-gray-400 font-normal text-sm">/hr</span></span>
          {host.rating > 0 && (
            <span className="text-yellow-400 text-sm">★ {host.rating.toFixed(1)}</span>
          )}
        </div>
      </div>
    </Link>
  )
}
