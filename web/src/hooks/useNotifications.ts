import { useState, useEffect, useCallback, useRef } from 'react'
import type { WSStateChangedPayload } from '@/types'

export interface UseNotificationsOptions {
  lastEvent?: WSStateChangedPayload | null
}

export interface UseNotificationsReturn {
  permission: NotificationPermission
  requestPermission: () => Promise<NotificationPermission>
}

/**
 * useNotifications 浏览器桌面交互通知 Hook
 */
export function useNotifications(options: UseNotificationsOptions = {}): UseNotificationsReturn {
  const { lastEvent } = options
  const [permission, setPermission] = useState<NotificationPermission>(() => {
    if (typeof window !== 'undefined' && 'Notification' in window) {
      return Notification.permission
    }
    return 'default'
  })

  const lastNotifiedEventRef = useRef<string | null>(null)

  const requestPermission = useCallback(async () => {
    if (typeof window !== 'undefined' && 'Notification' in window) {
      try {
        const result = await Notification.requestPermission()
        setPermission(result)
        return result
      } catch {
        return 'denied' as NotificationPermission
      }
    }
    return 'denied' as NotificationPermission
  }, [])

  // 监听状态变迁，当跃迁至 waiting_input 或 error 时触发桌面通知
  useEffect(() => {
    if (!lastEvent || permission !== 'granted') return

    const eventKey = `${lastEvent.pane_id}-${lastEvent.to_state}-${lastEvent.summary}`
    if (lastNotifiedEventRef.current === eventKey) {
      return
    }

    if (lastEvent.to_state === 'waiting_input' || lastEvent.to_state === 'error') {
      lastNotifiedEventRef.current = eventKey
      const isWaiting = lastEvent.to_state === 'waiting_input'
      const title = isWaiting
        ? `[Action Required] Pane ${lastEvent.pane_id} Waiting Input`
        : `[Agent Error] Pane ${lastEvent.pane_id} Error Detected`

      try {
        new Notification(title, {
          body: lastEvent.summary || `Agent transitioned from ${lastEvent.from_state} to ${lastEvent.to_state}`,
          tag: lastEvent.pane_id,
        })
      } catch {
        // 容错处理
      }
    }
  }, [lastEvent, permission])

  return {
    permission,
    requestPermission,
  }
}

