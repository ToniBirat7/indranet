const TOKEN_KEY = 'indranet_token'

export function setAuth(token: string) {
  if (typeof window === 'undefined') return
  localStorage.setItem(TOKEN_KEY, token)
}

export function getToken(): string | null {
  if (typeof window === 'undefined') return null
  return localStorage.getItem(TOKEN_KEY)
}

export function clearAuth() {
  if (typeof window === 'undefined') return
  localStorage.removeItem(TOKEN_KEY)
}
