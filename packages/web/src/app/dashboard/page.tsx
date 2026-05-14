// Server Component — user dashboard
export default async function DashboardPage() {
  // TODO: Fetch user sessions and wallet balance (server-side auth required)

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold text-white mb-8">Your Dashboard</h1>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="bg-gray-900 rounded-lg p-6">
          <p className="text-gray-400 text-sm">Wallet Balance</p>
          <p className="text-2xl font-bold text-white">$0.00</p>
          <button className="mt-3 text-brand-500 text-sm hover:text-brand-400">
            Add funds →
          </button>
        </div>
        <div className="bg-gray-900 rounded-lg p-6">
          <p className="text-gray-400 text-sm">Total Sessions</p>
          <p className="text-2xl font-bold text-white">0</p>
        </div>
        <div className="bg-gray-900 rounded-lg p-6">
          <p className="text-gray-400 text-sm">Hours Used</p>
          <p className="text-2xl font-bold text-white">0</p>
        </div>
      </div>

      <h2 className="text-xl font-semibold text-white mb-4">Recent Sessions</h2>
      <p className="text-gray-500">No sessions yet. Browse hosts to get started.</p>
    </div>
  )
}
