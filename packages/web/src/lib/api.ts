const API_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080'

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_URL}/v1${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    ...options,
  })

  if (!res.ok) {
    const error = await res.text()
    throw new Error(`API error ${res.status}: ${error}`)
  }

  return res.json()
}

export const api = {
  hosts: {
    list: (params?: Record<string, string>) => {
      const qs = params ? '?' + new URLSearchParams(params).toString() : ''
      return apiFetch<{ hosts: unknown[]; total: number }>(`/hosts${qs}`)
    },
    get: (id: string) => apiFetch<unknown>(`/hosts/${id}`),
  },
  sessions: {
    create: (body: { host_id: string; duration_minutes: number }, token: string) =>
      apiFetch<{ session_id: string; state: string; checkout_url?: string }>('/sessions', {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
        body: JSON.stringify(body),
      }),
    get: (id: string, token: string) =>
      apiFetch<unknown>(`/sessions/${id}`, {
        headers: { Authorization: `Bearer ${token}` },
      }),
    end: (id: string, token: string) =>
      apiFetch<{ state: string }>(`/sessions/${id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
      }),
  },
  auth: {
    login: (email: string, password: string) =>
      apiFetch<{ token: string }>('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      }),
    register: (email: string, password: string, name: string) =>
      apiFetch<{ user_id: string; token: string }>('/auth/register', {
        method: 'POST',
        body: JSON.stringify({ email, password, name }),
      }),
  },
}
