import { Suspense } from 'react'
import HostGrid from '@/components/marketplace/HostGrid'
import FilterBar from '@/components/marketplace/FilterBar'

interface Props {
  searchParams?: Record<string, string>
}

export default function HomePage({ searchParams }: Props) {
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

      <Suspense fallback={<div className="h-14 mb-6" />}>
        <FilterBar />
      </Suspense>
      <HostGrid searchParams={searchParams} />
    </div>
  )
}
