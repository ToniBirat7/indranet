'use client'

import { useEffect, useRef } from 'react'

interface Props {
  dataChannel: RTCDataChannel | null
}

// InputCapture hooks keyboard and mouse events and sends them to the host via the WebRTC data channel.
// Rendered unconditionally alongside StreamViewer; the useEffect no-ops while dataChannel is null.
export default function InputCapture({ dataChannel }: Props) {
  const locked = useRef(false)

  useEffect(() => {
    if (!dataChannel) return

    const send = (msg: object) => {
      if (dataChannel.readyState === 'open') dataChannel.send(JSON.stringify(msg))
    }

    const handleKeyDown = (e: KeyboardEvent) => {
      e.preventDefault()
      send({ t: 'k', e: 'd', c: e.code, k: e.keyCode })
    }

    const handleKeyUp = (e: KeyboardEvent) => {
      e.preventDefault()
      send({ t: 'k', e: 'u', c: e.code, k: e.keyCode })
    }

    const handleMouseMove = (e: MouseEvent) => {
      if (!locked.current) return
      // movementX/Y are relative deltas — only valid when pointer is locked
      send({ t: 'm', dx: e.movementX, dy: e.movementY })
    }

    const handleMouseDown = (e: MouseEvent) => {
      send({ t: 'mb', e: 'd', b: e.button })
    }

    const handleMouseUp = (e: MouseEvent) => {
      send({ t: 'mb', e: 'u', b: e.button })
    }

    const handleWheel = (e: WheelEvent) => {
      e.preventDefault()
      send({ t: 'mw', d: e.deltaY })
    }

    let gamepollId: number | null = null
    const pollGamepads = () => {
      const pads = navigator.getGamepads()
      for (const pad of pads) {
        if (!pad) continue
        const bt = pad.buttons.reduce((acc, b, i) => acc | (b.pressed ? 1 << i : 0), 0)
        send({
          t: 'gp',
          lx: pad.axes[0] ?? 0,
          ly: pad.axes[1] ?? 0,
          rx: pad.axes[2] ?? 0,
          ry: pad.axes[3] ?? 0,
          bt,
          lt: pad.buttons[6]?.value ?? 0,
          rt: pad.buttons[7]?.value ?? 0,
        })
        break // send only first connected gamepad
      }
      gamepollId = requestAnimationFrame(pollGamepads)
    }

    const onGamepadConnected = () => { if (gamepollId === null) gamepollId = requestAnimationFrame(pollGamepads) }
    const onGamepadDisconnected = () => { if (gamepollId !== null) { cancelAnimationFrame(gamepollId); gamepollId = null } }

    document.addEventListener('keydown', handleKeyDown)
    document.addEventListener('keyup', handleKeyUp)
    document.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mousedown', handleMouseDown)
    document.addEventListener('mouseup', handleMouseUp)
    document.addEventListener('wheel', handleWheel, { passive: false })
    window.addEventListener('gamepadconnected', onGamepadConnected)
    window.addEventListener('gamepaddisconnected', onGamepadDisconnected)

    const handlePointerLockChange = () => {
      locked.current = document.pointerLockElement !== null
    }
    document.addEventListener('pointerlockchange', handlePointerLockChange)

    return () => {
      document.removeEventListener('keydown', handleKeyDown)
      document.removeEventListener('keyup', handleKeyUp)
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mousedown', handleMouseDown)
      document.removeEventListener('mouseup', handleMouseUp)
      document.removeEventListener('wheel', handleWheel)
      document.removeEventListener('pointerlockchange', handlePointerLockChange)
      window.removeEventListener('gamepadconnected', onGamepadConnected)
      window.removeEventListener('gamepaddisconnected', onGamepadDisconnected)
      if (gamepollId !== null) cancelAnimationFrame(gamepollId)
    }
  }, [dataChannel])

  return null // Invisible component, just hooks events
}
