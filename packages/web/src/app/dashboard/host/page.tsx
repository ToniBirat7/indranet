// Server Component — host earnings dashboard
export default async function HostDashboardPage() {
  // TODO: Fetch host's sessions, earnings, and payout history

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold text-white mb-8">Host Dashboard</h1>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="bg-gray-900 rounded-lg p-6">
          <p className="text-gray-400 text-sm">This Month Earnings</p>
          <p className="text-2xl font-bold text-green-400">$0.00</p>
        </div>
        <div className="bg-gray-900 rounded-lg p-6">
          <p className="text-gray-400 text-sm">Total Sessions</p>
          <p className="text-2xl font-bold text-white">0</p>
        </div>
        <div className="bg-gray-900 rounded-lg p-6">
          <p className="text-gray-400 text-sm">Status</p>
          <p className="text-2xl font-bold text-gray-500">Offline</p>
        </div>
      </div>

      <h2 className="text-xl font-semibold text-white mb-4">Recent Sessions</h2>
      <p className="text-gray-500">No sessions yet.</p>
    </div>
  )
}
