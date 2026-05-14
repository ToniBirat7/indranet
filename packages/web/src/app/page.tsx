import HostGrid from '@/components/marketplace/HostGrid'
import FilterBar from '@/components/marketplace/FilterBar'

export default function HomePage() {
  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-4xl font-bold text-white mb-2">
          Rent GPU Compute
        </h1>
        <p className="text-gray-400 text-lg">
          Live, interactive sessions on real gaming hardware. Pay by the minute.
        </p>
      </div>

      <FilterBar />
      <HostGrid />
    </div>
  )
}
