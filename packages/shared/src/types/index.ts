// Canonical TypeScript types — must stay in sync with Go models.
// Field names use snake_case to match the Go JSON API responses exactly.

export interface User {
  user_id: string
  email: string
  name: string
  role: string
  balance_cents: number
  host_id?: string
  created_at: string
}

export interface Host {
  host_id: string
  display_name: string
  gpu_model: string
  vram_gb: number
  cpu_model: string
  ram_gb: number
  os: string
  price_per_hour_cents: number
  price_per_minute_cents: number
  tags: string[]
  online: boolean
  rating: number
  total_sessions: number
  payouts_enabled?: boolean
}

export type SessionState = 'CREATED' | 'AUTHORIZED' | 'ACTIVE' | 'ENDING' | 'ENDED' | 'FAILED'

export interface Session {
  session_id: string
  host_id: string
  state: SessionState
  rate_per_minute_cents: number
  total_charged_cents: number
  balance_remaining_minutes?: number
  started_at: string | null
  created_at: string
}

// WebRTC signaling message types (relayed via the Go signaling hub)
export type SignalMessage =
  | { type: 'offer'; sdp: string }
  | { type: 'answer'; sdp: string }
  | { type: 'ice_candidate'; candidate: RTCIceCandidateInit }
  | { type: 'session_authorized'; session_id: string }
  | { type: 'session_state'; state: SessionState }
  | { type: 'session_kill'; reason: string }
  | { type: 'session_failed'; reason: string }
  | { type: 'session_warning'; minutes_remaining: number }

// Compact input event format for WebRTC data channel (ordered:false, maxRetransmits:0)
export type InputEvent =
  | { t: 'k'; e: 'd' | 'u'; c: string; k: number }          // keyboard down/up
  | { t: 'm'; dx: number; dy: number }                        // mouse move (pointer-locked)
  | { t: 'mb'; e: 'd' | 'u'; b: number }                     // mouse button down/up
  | { t: 'mw'; d: number }                                    // mouse wheel
  | { t: 'gp'; lx: number; ly: number; rx: number; ry: number; bt: number; lt: number; rt: number }
