import HostCard from './HostCard'

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

// Server Component — fetches hosts from the Go API on the server.
export default async function HostGrid() {
  const apiURL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080'

  let hosts: Host[] = []
  try {
    const res = await fetch(`${apiURL}/v1/hosts`, { cache: 'no-store' })
    if (res.ok) {
      const data = await res.json()
      hosts = data.hosts ?? []
    }
  } catch {
    // Backend unreachable — show empty state, not an error
  }

  if (!hosts.length) {
    return (
      <div className="text-center py-16">
        <p className="text-gray-500 text-lg">No hosts available yet.</p>
        <p className="text-gray-600 text-sm mt-2">
          Check back soon, or{' '}
          <a href="/dashboard/host" className="text-brand-400 hover:text-brand-300">
            become a host
          </a>
          .
        </p>
      </div>
    )
  }

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {hosts.map((host) => (
        <HostCard key={host.host_id} host={host} />
      ))}
    </div>
  )
}
