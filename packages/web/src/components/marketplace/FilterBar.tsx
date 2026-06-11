'use client'

import { useRouter, useSearchParams, usePathname } from 'next/navigation'
import { useCallback } from 'react'

export default function FilterBar() {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()

  const setParam = useCallback(
    (key: string, value: string) => {
      const params = new URLSearchParams(searchParams.toString())
      if (value) {
        params.set(key, value)
      } else {
        params.delete(key)
      }
      router.push(`${pathname}?${params.toString()}`)
    },
    [router, pathname, searchParams],
  )

  return (
    <div className="flex flex-wrap gap-4 mb-6 p-4 bg-gray-900 rounded-xl border border-gray-800">
      <div>
        <label className="text-gray-400 text-sm block mb-1">Min VRAM (GB)</label>
        <select
          value={searchParams.get('min_vram') ?? ''}
          onChange={(e) => setParam('min_vram', e.target.value)}
          className="bg-gray-800 text-white border border-gray-700 rounded px-3 py-1.5 text-sm"
        >
          <option value="">Any</option>
          <option value="8">8 GB+</option>
          <option value="16">16 GB+</option>
          <option value="24">24 GB+</option>
        </select>
      </div>
      <div>
        <label className="text-gray-400 text-sm block mb-1">Max Price</label>
        <select
          value={searchParams.get('max_price_cents') ?? ''}
          onChange={(e) => setParam('max_price_cents', e.target.value)}
          className="bg-gray-800 text-white border border-gray-700 rounded px-3 py-1.5 text-sm"
        >
          <option value="">Any</option>
          <option value="100">Under $1/hr</option>
          <option value="200">Under $2/hr</option>
          <option value="500">Under $5/hr</option>
        </select>
      </div>
      <div className="flex items-end">
        <label className="flex items-center gap-2 text-sm text-gray-300 cursor-pointer">
          <input
            type="checkbox"
            checked={searchParams.get('online') === '1'}
            onChange={(e) => setParam('online', e.target.checked ? '1' : '')}
            className="accent-brand-500"
          />
          Online only
        </label>
      </div>
      {(searchParams.get('min_vram') || searchParams.get('max_price_cents') || searchParams.get('online')) && (
        <div className="flex items-end">
          <button
            onClick={() => router.push(pathname)}
            className="text-gray-500 hover:text-gray-300 text-sm"
          >
            Clear filters
          </button>
        </div>
      )}
    </div>
  )
}
