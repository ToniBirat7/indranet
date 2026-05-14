import HostCard from './HostCard'

// Server Component — fetches hosts from the API on the server.
export default async function HostGrid() {
  // TODO: Fetch hosts from backend API
  // const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/v1/hosts`, { cache: 'no-store' })
  // const data = await res.json()
  // const hosts = data.hosts

  return (
    <div>
      <p className="text-gray-500 text-center py-16">
        No hosts available yet. Check back soon!
      </p>
    </div>
  )
}
