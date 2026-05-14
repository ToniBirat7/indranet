// WebRTC session utilities
// Used by StreamViewer and InputCapture components

export interface WebRTCSession {
  pc: RTCPeerConnection
  ws: WebSocket
  inputChannel: RTCDataChannel | null
}

export interface WebRTCConfig {
  signalingUrl: string
  iceServers?: RTCIceServer[]
}

const DEFAULT_ICE_SERVERS: RTCIceServer[] = [
  { urls: 'stun:stun.l.google.com:19302' },
  // TODO: Add TURN server from env
]

// createWebRTCSession creates a peer connection and returns it along with the signaling WebSocket.
// The caller is responsible for calling pc.close() and ws.close() on cleanup.
export async function createWebRTCSession(config: WebRTCConfig): Promise<WebRTCSession> {
  const pc = new RTCPeerConnection({
    iceServers: config.iceServers ?? DEFAULT_ICE_SERVERS,
  })

  const ws = new WebSocket(config.signalingUrl)

  // TODO: Open input data channel
  // const inputChannel = pc.createDataChannel('input', {
  //   ordered: false,     // UDP-like — don't wait for retransmission
  //   maxRetransmits: 0,  // Drop instead of retransmit (input events are time-sensitive)
  // })

  return new Promise((resolve, reject) => {
    ws.onopen = () => resolve({ pc, ws, inputChannel: null })
    ws.onerror = () => reject(new Error('WebSocket connection failed'))
  })
}
