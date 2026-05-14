'use client'

import { useEffect, useRef } from 'react'

interface Props {
  dataChannel: RTCDataChannel | null
}

// InputCapture hooks keyboard and mouse events and sends them to the host via the WebRTC data channel.
// Must be rendered as a sibling to StreamViewer after the WebRTC connection is established.
// TODO: Mount this component once the data channel is open.
export default function InputCapture({ dataChannel }: Props) {
  const locked = useRef(false)

  useEffect(() => {
    if (!dataChannel) return

    const handleKeyDown = (e: KeyboardEvent) => {
      e.preventDefault()
      dataChannel.send(JSON.stringify({ t: 'k', e: 'd', c: e.code, k: e.keyCode }))
    }

    const handleKeyUp = (e: KeyboardEvent) => {
      e.preventDefault()
      dataChannel.send(JSON.stringify({ t: 'k', e: 'u', c: e.code, k: e.keyCode }))
    }

    const handleMouseMove = (e: MouseEvent) => {
      if (!locked.current) return
      // movementX/Y are relative deltas — only valid when pointer is locked
      dataChannel.send(JSON.stringify({ t: 'm', dx: e.movementX, dy: e.movementY }))
    }

    const handleMouseDown = (e: MouseEvent) => {
      dataChannel.send(JSON.stringify({ t: 'mb', e: 'd', b: e.button }))
    }

    const handleMouseUp = (e: MouseEvent) => {
      dataChannel.send(JSON.stringify({ t: 'mb', e: 'u', b: e.button }))
    }

    document.addEventListener('keydown', handleKeyDown)
    document.addEventListener('keyup', handleKeyUp)
    document.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mousedown', handleMouseDown)
    document.addEventListener('mouseup', handleMouseUp)

    // Track pointer lock state
    document.addEventListener('pointerlockchange', () => {
      locked.current = document.pointerLockElement !== null
    })

    return () => {
      document.removeEventListener('keydown', handleKeyDown)
      document.removeEventListener('keyup', handleKeyUp)
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mousedown', handleMouseDown)
      document.removeEventListener('mouseup', handleMouseUp)
    }
  }, [dataChannel])

  return null // Invisible component, just hooks events
}
