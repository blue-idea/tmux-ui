import { useState, useEffect, useRef, useCallback } from 'react'
import type {
  PaneSummary,
  ConnectionStatus,
  WSMessage,
  WSSnapshotPayload,
  WSStateChangedPayload,
} from '@/types'

export interface UseTmuxWSOptions {
  url?: string
  token?: string
  reconnectInterval?: number
}

export interface UseTmuxWSReturn {
  status: ConnectionStatus
  panes: PaneSummary[]
  lastEvent: WSStateChangedPayload | null
  reconnect: () => void
}

/**
 * useTmuxWS WebSocket 实时通信与状态同步 Hook
 */
export function useTmuxWS(options: UseTmuxWSOptions = {}): UseTmuxWSReturn {
  const { url, token, reconnectInterval = 2000 } = options

  const [status, setStatus] = useState<ConnectionStatus>('connecting')
  const [panes, setPanes] = useState<PaneSummary[]>([])
  const [lastEvent, setLastEvent] = useState<WSStateChangedPayload | null>(null)

  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const isUnmountedRef = useRef(false)

  // 构造 WebSocket 连接地址
  const buildWSUrl = useCallback(() => {
    let targetUrl = url
    if (!targetUrl) {
      if (typeof window !== 'undefined' && window.location) {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
        targetUrl = `${protocol}//${window.location.host}/ws`
      } else {
        targetUrl = 'ws://127.0.0.1:8080/ws'
      }
    }

    if (token) {
      const separator = targetUrl.includes('?') ? '&' : '?'
      targetUrl = `${targetUrl}${separator}token=${encodeURIComponent(token)}`
    }
    return targetUrl
  }, [url, token])

  const connect = useCallback(() => {
    if (isUnmountedRef.current) return

    // 清理之前的连接与定时器
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    if (reconnectTimerRef.current) {
      clearTimeout(reconnectTimerRef.current)
      reconnectTimerRef.current = null
    }

    setStatus('connecting')

    try {
      const fullUrl = buildWSUrl()
      const ws = new WebSocket(fullUrl)
      wsRef.current = ws

      ws.onopen = () => {
        if (isUnmountedRef.current) return
        setStatus('connected')
      }

      ws.onmessage = (event: MessageEvent) => {
        if (isUnmountedRef.current) return
        try {
          const msg = JSON.parse(event.data) as WSMessage<unknown>
          switch (msg.type) {
            case 'snapshot': {
              const payload = msg.payload as WSSnapshotPayload
              if (payload && Array.isArray(payload.panes)) {
                setPanes(payload.panes)
              }
              break
            }

            case 'state_changed': {
              const payload = msg.payload as WSStateChangedPayload
              if (payload && payload.pane_id) {
                setLastEvent(payload)
                setPanes((prev) =>
                  prev.map((p) =>
                    p.pane_id === payload.pane_id
                      ? { ...p, state: payload.to_state }
                      : p
                  )
                )
              }
              break
            }

            case 'output_diff': {
              const payload = msg.payload as { pane_id: string; tail_preview: string[] }
              if (payload && payload.pane_id && Array.isArray(payload.tail_preview)) {
                setPanes((prev) =>
                  prev.map((p) =>
                    p.pane_id === payload.pane_id
                      ? { ...p, tail_preview: payload.tail_preview, idle_seconds: 0 }
                      : p
                  )
                )
              }
              break
            }

            default:
              break
          }
        } catch {
          // 忽略格式异常消息
        }
      }

      ws.onerror = () => {
        if (isUnmountedRef.current) return
        setStatus('disconnected')
      }

      ws.onclose = () => {
        if (isUnmountedRef.current) return
        setStatus('disconnected')
        // 自动调度重连
        reconnectTimerRef.current = setTimeout(() => {
          connect()
        }, reconnectInterval)
      }
    } catch {
      setStatus('disconnected')
      reconnectTimerRef.current = setTimeout(() => {
        connect()
      }, reconnectInterval)
    }
  }, [buildWSUrl, reconnectInterval])

  useEffect(() => {
    isUnmountedRef.current = false
    connect()

    return () => {
      isUnmountedRef.current = true
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current)
      }
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [connect])

  const reconnect = useCallback(() => {
    connect()
  }, [connect])

  return {
    status,
    panes,
    lastEvent,
    reconnect,
  }
}

