'use client'

// FilterBar is a Client Component because it manages filter state interactively.
export default function FilterBar() {
  return (
    <div className="flex flex-wrap gap-4 mb-6 p-4 bg-gray-900 rounded-xl border border-gray-800">
      <div>
        <label className="text-gray-400 text-sm block mb-1">Min VRAM (GB)</label>
        <select className="bg-gray-800 text-white border border-gray-700 rounded px-3 py-1.5 text-sm">
          <option value="">Any</option>
          <option value="8">8 GB+</option>
          <option value="16">16 GB+</option>
          <option value="24">24 GB+</option>
        </select>
      </div>
      <div>
        <label className="text-gray-400 text-sm block mb-1">Max Price</label>
        <select className="bg-gray-800 text-white border border-gray-700 rounded px-3 py-1.5 text-sm">
          <option value="">Any</option>
          <option value="1">Under $1/hr</option>
          <option value="2">Under $2/hr</option>
          <option value="5">Under $5/hr</option>
        </select>
      </div>
      <div className="flex items-end">
        <label className="flex items-center gap-2 text-sm text-gray-300 cursor-pointer">
          <input type="checkbox" className="accent-brand-500" />
          Online only
        </label>
      </div>
    </div>
  )
}
