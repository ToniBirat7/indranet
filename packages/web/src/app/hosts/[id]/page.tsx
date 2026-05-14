interface Props {
  params: { id: string }
}

// Server Component — fetch host data on the server
export default async function HostDetailPage({ params }: Props) {
  // TODO: fetch(`${process.env.NEXT_PUBLIC_API_URL}/v1/hosts/${params.id}`)

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold text-white mb-4">Host Details</h1>
      <p className="text-gray-400">Host ID: {params.id}</p>
      {/* TODO: Render host specs, pricing, and Book Session form */}
    </div>
  )
}
