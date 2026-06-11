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

export interface UserProfile {
  user_id: string
  email: string
  name: string
  role: string
  balance_cents: number
  host_id?: string
}

export interface SessionSummary {
  session_id: string
  host_id: string
  host_name: string
  state: string
  rate_per_minute_cents: number
  total_charged_cents: number
  started_at: string | null
  created_at: string
}

export const api = {
  hosts: {
    list: (params?: Record<string, string>) => {
      const qs = params ? '?' + new URLSearchParams(params).toString() : ''
      return apiFetch<{ hosts: unknown[]; total: number }>(`/hosts${qs}`)
    },
    get: (id: string) => apiFetch<unknown>(`/hosts/${id}`),
    register: (
      body: {
        display_name: string
        gpu_model: string
        vram_gb: number
        cpu_model: string
        ram_gb: number
        os: string
        price_per_hour_cents: number
        tags: string[]
      },
      token: string,
    ) =>
      apiFetch<{ host_id: string; agent_token: string }>('/hosts/register', {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
        body: JSON.stringify(body),
      }),
    connectStripe: (token: string) =>
      apiFetch<{ onboarding_url: string }>('/hosts/me/stripe/connect', {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
      }),
  },
  sessions: {
    list: (token: string, params?: { page?: number; limit?: number }) => {
      const qs = params ? '?' + new URLSearchParams(Object.entries(params).map(([k, v]) => [k, String(v)])).toString() : ''
      return apiFetch<{ sessions: SessionSummary[]; total: number; page: number; limit: number }>(`/sessions${qs}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
    },
    create: (body: { host_id: string; duration_minutes: number }, token: string) =>
      apiFetch<{ session_id: string; state: string; checkout_url?: string }>('/sessions', {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
        body: JSON.stringify(body),
      }),
    get: (id: string, token: string) =>
      apiFetch<{
        session_id: string
        state: string
        rate_per_minute_cents: number
        balance_remaining_minutes: number
        total_charged_cents: number
      }>(`/sessions/${id}`, {
        headers: { Authorization: `Bearer ${token}` },
      }),
    end: (id: string, token: string) =>
      apiFetch<{ state: string }>(`/sessions/${id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
      }),
    rate: (id: string, stars: number, token: string) =>
      apiFetch<{ stars: number }>(`/sessions/${id}/rate`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
        body: JSON.stringify({ stars }),
      }),
  },
  users: {
    me: (token: string) =>
      apiFetch<UserProfile>('/users/me', {
        headers: { Authorization: `Bearer ${token}` },
      }),
    topup: (amountCents: number, token: string) =>
      apiFetch<{ checkout_url?: string; credited_cents?: number; dev_mode?: boolean }>('/users/me/topup', {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
        body: JSON.stringify({ amount_cents: amountCents }),
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
