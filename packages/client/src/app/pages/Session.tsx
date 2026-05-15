'use client'

import { useEffect, useRef, useState } from 'react'

interface Props {
  sessionId: string
  signalingUrl: string
  token: string
}

export default function Session({ sessionId, signalingUrl, token }: Props) {
  const videoRef = useRef<HTMLVideoElement>(null)
  const [status, setStatus] = useState<'connecting' | 'connected' | 'error'>('connecting')

  useEffect(() => {
    let pc: RTCPeerConnection | null = null
    let ws: WebSocket | null = null

    async function connect() {
      pc = new RTCPeerConnection({
        iceServers: [{ urls: 'stun:stun.l.google.com:19302' }],
      })

      pc.ontrack = (event) => {
        if (videoRef.current) {
          videoRef.current.srcObject = event.streams[0]
        }
        setStatus('connected')
      }

      ws = new WebSocket(`${signalingUrl}?session=${sessionId}&token=${token}`)
      ws.onmessage = async (event) => {
        const msg = JSON.parse(event.data)
        if (msg.type === 'offer') {
          await pc!.setRemoteDescription({ type: 'offer', sdp: msg.sdp })
          const answer = await pc!.createAnswer()
          await pc!.setLocalDescription(answer)
          ws!.send(JSON.stringify({ type: 'answer', sdp: answer.sdp }))
        } else if (msg.type === 'ice_candidate') {
          await pc!.addIceCandidate(msg.candidate)
        }
      }
      ws.onerror = () => setStatus('error')

      pc.onicecandidate = (event) => {
        if (event.candidate && ws?.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ type: 'ice_candidate', candidate: event.candidate }))
        }
      }
    }

    connect()
    return () => {
      pc?.close()
      ws?.close()
    }
  }, [sessionId, signalingUrl, token])

  return (
    <div className="flex flex-col h-screen bg-black">
      {status !== 'connected' && (
        <div className="absolute inset-0 flex items-center justify-center text-white">
          {status === 'connecting' ? 'Connecting...' : 'Connection error'}
        </div>
      )}
      <video
        ref={videoRef}
        autoPlay
        playsInline
        className="w-full h-full object-contain"
      />
    </div>
  )
}
