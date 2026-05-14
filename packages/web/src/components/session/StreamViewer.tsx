'use client'

import { useEffect, useRef, useState } from 'react'

interface Props {
  sessionId: string
  signalingUrl: string
}

type ConnectionState = 'idle' | 'connecting' | 'connected' | 'error'

// StreamViewer handles the WebRTC peer connection and renders the remote video stream.
// It also opens a data channel for sending keyboard/mouse/gamepad input to the host.
export default function StreamViewer({ sessionId, signalingUrl }: Props) {
  const videoRef = useRef<HTMLVideoElement>(null)
  const pcRef = useRef<RTCPeerConnection | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const [state, setState] = useState<ConnectionState>('idle')
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    connect()
    return () => disconnect()
  }, [sessionId, signalingUrl])

  async function connect() {
    setState('connecting')
    setError(null)

    // TODO: Connect to signaling WebSocket
    const wsUrl = `${signalingUrl}?role=viewer&token=${getSessionToken()}`
    const ws = new WebSocket(wsUrl)
    wsRef.current = ws

    ws.onopen = () => {
      setupPeerConnection(ws)
    }

    ws.onerror = () => {
      setError('Failed to connect to signaling server')
      setState('error')
    }

    ws.onclose = () => {
      if (state === 'connected') {
        setError('Connection lost')
        setState('error')
      }
    }
  }

  function setupPeerConnection(ws: WebSocket) {
    const pc = new RTCPeerConnection({
      iceServers: [
        { urls: 'stun:stun.l.google.com:19302' },
        // TODO: Add TURN server config from environment
      ],
    })
    pcRef.current = pc

    // Handle incoming media tracks (video + audio from host)
    pc.ontrack = (event) => {
      if (videoRef.current && event.streams[0]) {
        videoRef.current.srcObject = event.streams[0]
        setState('connected')
      }
    }

    // ICE candidate relay via signaling server
    pc.onicecandidate = (event) => {
      if (event.candidate) {
        ws.send(JSON.stringify({ type: 'ice_candidate', candidate: event.candidate }))
      }
    }

    pc.onconnectionstatechange = () => {
      if (pc.connectionState === 'failed') {
        setError('WebRTC connection failed')
        setState('error')
      }
    }

    // TODO: Open data channel for input events
    // const inputChannel = pc.createDataChannel('input', { ordered: false })
    // inputChannel.onopen = () => startInputCapture(inputChannel)

    // Handle incoming SDP offer from host agent (relayed via signaling server)
    ws.onmessage = async (event) => {
      const msg = JSON.parse(event.data)

      if (msg.type === 'offer') {
        await pc.setRemoteDescription(new RTCSessionDescription({ type: 'offer', sdp: msg.sdp }))
        const answer = await pc.createAnswer()
        await pc.setLocalDescription(answer)
        ws.send(JSON.stringify({ type: 'answer', sdp: answer.sdp }))
      } else if (msg.type === 'ice_candidate') {
        await pc.addIceCandidate(new RTCIceCandidate(msg.candidate))
      } else if (msg.type === 'session_kill') {
        // Billing exhausted or host terminated session
        setError('Session ended by host or balance exhausted')
        setState('error')
      }
    }
  }

  function disconnect() {
    pcRef.current?.close()
    wsRef.current?.close()
    if (videoRef.current) {
      videoRef.current.srcObject = null
    }
  }

  function getSessionToken(): string {
    // TODO: Get JWT from auth context or localStorage
    return ''
  }

  if (state === 'error') {
    return (
      <div className="flex items-center justify-center h-full bg-gray-950 text-white">
        <div className="text-center">
          <p className="text-red-400 text-lg mb-4">{error ?? 'Connection error'}</p>
          <button
            onClick={connect}
            className="bg-brand-600 hover:bg-brand-700 px-6 py-2 rounded-lg"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  if (state === 'connecting') {
    return (
      <div className="flex items-center justify-center h-full bg-gray-950 text-white">
        <div className="text-center">
          <div className="animate-spin w-8 h-8 border-2 border-brand-500 border-t-transparent rounded-full mx-auto mb-4" />
          <p className="text-gray-400">Connecting to host...</p>
        </div>
      </div>
    )
  }

  return (
    <video
      ref={videoRef}
      autoPlay
      playsInline
      className="w-full h-full object-contain bg-black cursor-none"
      // TODO: Attach pointer lock + keyboard/mouse event listeners
    />
  )
}
