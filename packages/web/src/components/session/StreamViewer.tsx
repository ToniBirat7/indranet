'use client'

import { useEffect, useRef, useState } from 'react'
import { getToken } from '@/lib/auth'
import InputCapture from './InputCapture'

type SessionEvent = {
  type: string
  reason?: string
  minutes_remaining?: number
}

interface Props {
  sessionId: string
  signalingUrl: string
  onSessionEvent?: (event: SessionEvent) => void
}

type ConnectionState = 'idle' | 'connecting' | 'connected' | 'error'

export default function StreamViewer({ sessionId, signalingUrl, onSessionEvent }: Props) {
  const videoRef = useRef<HTMLVideoElement>(null)
  const pcRef = useRef<RTCPeerConnection | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const [state, setState] = useState<ConnectionState>('idle')
  const [error, setError] = useState<string | null>(null)
  const [inputChannel, setInputChannel] = useState<RTCDataChannel | null>(null)

  useEffect(() => {
    connect()
    return () => disconnect()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sessionId, signalingUrl])

  async function connect() {
    setState('connecting')
    setError(null)

    const token = getToken() ?? ''
    const wsUrl = `${signalingUrl}?role=viewer${token ? `&token=${token}` : ''}`
    const ws = new WebSocket(wsUrl)
    wsRef.current = ws

    ws.onopen = () => setupPeerConnection(ws)
    ws.onerror = () => { setError('Failed to connect to signaling server'); setState('error') }
    ws.onclose = () => setState((prev) => (prev === 'connected' || prev === 'connecting' ? 'error' : prev))
  }

  function buildIceServers(): RTCIceServer[] {
    const servers: RTCIceServer[] = [{ urls: 'stun:stun.l.google.com:19302' }]
    const turnUrl = process.env.NEXT_PUBLIC_TURN_URL
    if (turnUrl) {
      servers.push({
        urls: turnUrl,
        username: process.env.NEXT_PUBLIC_TURN_USERNAME ?? '',
        credential: process.env.NEXT_PUBLIC_TURN_CREDENTIAL ?? '',
      })
    }
    return servers
  }

  function setupPeerConnection(ws: WebSocket) {
    const pc = new RTCPeerConnection({ iceServers: buildIceServers() })
    pcRef.current = pc

    // Input data channel — ordered: false, maxRetransmits: 0 for minimal latency
    const ch = pc.createDataChannel('input', { ordered: false, maxRetransmits: 0 })
    ch.onopen = () => setInputChannel(ch)
    ch.onclose = () => setInputChannel(null)

    pc.ontrack = (event) => {
      if (videoRef.current && event.streams[0]) {
        videoRef.current.srcObject = event.streams[0]
        setState('connected')
      }
    }

    pc.onicecandidate = (event) => {
      if (event.candidate) ws.send(JSON.stringify({ type: 'ice_candidate', candidate: event.candidate }))
    }

    pc.onconnectionstatechange = () => {
      if (pc.connectionState === 'failed') { setError('WebRTC connection failed'); setState('error') }
    }

    ws.onmessage = async (rawEvent) => {
      let msg: Record<string, unknown>
      try { msg = JSON.parse(rawEvent.data as string) } catch { return }

      switch (msg.type) {
        case 'offer': {
          await pc.setRemoteDescription({ type: 'offer', sdp: msg.sdp as string })
          const answer = await pc.createAnswer()
          await pc.setLocalDescription(answer)
          ws.send(JSON.stringify({ type: 'answer', sdp: answer.sdp }))
          break
        }
        case 'ice_candidate':
          if (msg.candidate) await pc.addIceCandidate(new RTCIceCandidate(msg.candidate as RTCIceCandidateInit))
          break
        case 'session_kill':
        case 'session_failed':
        case 'session_warning':
          onSessionEvent?.(msg as SessionEvent)
          break
      }
    }
  }

  function disconnect() {
    pcRef.current?.close()
    wsRef.current?.close()
    if (videoRef.current) videoRef.current.srcObject = null
    setInputChannel(null)
  }

  if (state === 'error') {
    return (
      <div className="flex items-center justify-center h-full bg-gray-950 text-white">
        <div className="text-center">
          <p className="text-red-400 text-lg mb-4">{error ?? 'Connection error'}</p>
          <button onClick={() => { disconnect(); connect() }} className="bg-brand-600 hover:bg-brand-700 px-6 py-2 rounded-lg">Retry</button>
        </div>
      </div>
    )
  }

  if (state !== 'connected') {
    return (
      <div className="flex items-center justify-center h-full bg-gray-950 text-white">
        <div className="text-center">
          <div className="animate-spin w-8 h-8 border-2 border-brand-500 border-t-transparent rounded-full mx-auto mb-4" />
          <p className="text-gray-400">Connecting to host…</p>
        </div>
      </div>
    )
  }

  return (
    <>
      <video
        ref={videoRef}
        autoPlay
        playsInline
        className="w-full h-full object-contain bg-black cursor-none"
        onClick={() => videoRef.current?.requestPointerLock()}
        title="Click to capture mouse input"
      />
      <InputCapture dataChannel={inputChannel} />
    </>
  )
}
