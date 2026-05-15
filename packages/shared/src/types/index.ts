export interface User {
  id: string
  email: string
  name: string
  walletBalanceCents: number
  createdAt: string
}

export interface Host {
  id: string
  ownerId: string
  name: string
  description: string
  gpuModel: string
  ramGb: number
  cpuCores: number
  pricePerMinuteCents: number
  isOnline: boolean
  rating: number
  totalSessions: number
  createdAt: string
}

export type SessionState =
  | 'created'
  | 'authorized'
  | 'active'
  | 'ending'
  | 'ended'
  | 'failed'

export interface Session {
  id: string
  hostId: string
  userId: string
  state: SessionState
  durationMinutes: number
  totalCostCents: number
  startedAt: string | null
  endedAt: string | null
  createdAt: string
}

export interface BillingInfo {
  sessionId: string
  secondsElapsed: number
  secondsRemaining: number
  totalCostCents: number
  action: 'continue' | 'warn' | 'kill'
}

// WebRTC signaling message types
export type SignalMessage =
  | { type: 'offer'; sdp: string }
  | { type: 'answer'; sdp: string }
  | { type: 'ice_candidate'; candidate: RTCIceCandidateInit }
  | { type: 'session_state'; state: SessionState }
  | { type: 'billing_update'; billing: BillingInfo }

// Input event types sent over WebRTC data channel
export type InputEvent =
  | { type: 'keyboard'; action: 'down' | 'up'; key: number; modifiers: number }
  | { type: 'mouse'; dx: number; dy: number; buttons: number; wheel: number }
  | { type: 'gamepad'; leftX: number; leftY: number; rightX: number; rightY: number; buttons: number; leftTrigger: number; rightTrigger: number }
