'use client'

import Link from 'next/link'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { getToken, clearAuth } from '@/lib/auth'

export default function Navbar() {
  const router = useRouter()
  const [loggedIn, setLoggedIn] = useState(false)

  // Read auth state after hydration (localStorage unavailable during SSR)
  useEffect(() => {
    setLoggedIn(!!getToken())
  }, [])

  function handleSignOut() {
    clearAuth()
    setLoggedIn(false)
    router.push('/')
    router.refresh()
  }

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
          {loggedIn ? (
            <>
              <Link href="/dashboard" className="text-gray-400 hover:text-white text-sm transition-colors">
                Dashboard
              </Link>
              <button
                onClick={handleSignOut}
                className="text-gray-400 hover:text-white text-sm transition-colors"
              >
                Sign out
              </button>
            </>
          ) : (
            <Link
              href="/auth/login"
              className="bg-brand-600 hover:bg-brand-700 text-white text-sm px-4 py-2 rounded-lg transition-colors"
            >
              Sign In
            </Link>
          )}
        </div>
      </div>
    </nav>
  )
}
