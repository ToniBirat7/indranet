// Re-export shared types from @indranet/shared
// Add web-specific types below

export type { Host, Session, User } from '@indranet/shared'

export interface SessionViewerState {
  connected: boolean
  loading: boolean
  error: string | null
}
