import Link from 'next/link'

export default function Navbar() {
  return (
    <nav className="border-b border-gray-800 bg-gray-950/80 backdrop-blur-sm sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 h-16 flex items-center justify-between">
        <Link href="/" className="text-xl font-bold text-brand-400 tracking-tight">
          IndraNet
        </Link>

        <div className="flex items-center gap-6">
          <Link href="/" className="text-gray-400 hover:text-white text-sm transition-colors">
            Browse
          </Link>
          <Link href="/dashboard/host" className="text-gray-400 hover:text-white text-sm transition-colors">
            Become a Host
          </Link>
          <Link href="/dashboard" className="text-gray-400 hover:text-white text-sm transition-colors">
            Dashboard
          </Link>
          <Link
            href="/auth/login"
            className="bg-brand-600 hover:bg-brand-700 text-white text-sm px-4 py-2 rounded-lg transition-colors"
          >
            Sign In
          </Link>
        </div>
      </div>
    </nav>
  )
}
