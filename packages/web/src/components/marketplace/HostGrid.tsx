import Link from 'next/link'
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

interface Props {
  searchParams?: Record<string, string>
}

const PAGE_SIZE = 20

// Server Component — fetches hosts from the Go API on the server.
export default async function HostGrid({ searchParams }: Props) {
  const apiURL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080'

  const page = Math.max(1, parseInt(searchParams?.page ?? '1', 10) || 1)
  const params = new URLSearchParams()
  if (searchParams?.min_vram) params.set('min_vram', searchParams.min_vram)
  if (searchParams?.max_price_cents) params.set('max_price_cents', searchParams.max_price_cents)
  if (searchParams?.online === '1') params.set('online', '1')
  params.set('page', String(page))
  params.set('limit', String(PAGE_SIZE))

  let hosts: Host[] = []
  let total = 0
  try {
    const res = await fetch(`${apiURL}/v1/hosts?${params.toString()}`, { cache: 'no-store' })
    if (res.ok) {
      const data = await res.json()
      hosts = data.hosts ?? []
      total = data.total ?? 0
    }
  } catch {
    // Backend unreachable — show empty state, not an error
  }

  if (!hosts.length) {
    return (
      <div className="text-center py-16">
        <p className="text-gray-500 text-lg">No hosts available.</p>
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

  const totalPages = Math.ceil(total / PAGE_SIZE)

  const buildPageUrl = (p: number) => {
    const sp = new URLSearchParams(searchParams ?? {})
    sp.set('page', String(p))
    return `/?${sp.toString()}`
  }

  return (
    <>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mb-8">
        {hosts.map((host) => (
          <HostCard key={host.host_id} host={host} />
        ))}
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          {page > 1 && (
            <Link href={buildPageUrl(page - 1)}
              className="px-4 py-2 rounded-lg bg-gray-800 text-gray-300 hover:bg-gray-700 text-sm">
              ← Prev
            </Link>
          )}
          <span className="text-gray-500 text-sm">
            Page {page} of {totalPages}
          </span>
          {page < totalPages && (
            <Link href={buildPageUrl(page + 1)}
              className="px-4 py-2 rounded-lg bg-gray-800 text-gray-300 hover:bg-gray-700 text-sm">
              Next →
            </Link>
          )}
        </div>
      )}
    </>
  )
}
