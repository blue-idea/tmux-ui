import { renderHook, act } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useNotifications } from './useNotifications'
import type { WSStateChangedPayload } from '@/types'

// Mock 浏览器 Notification API
const mockNotificationInstance = vi.fn()
class MockNotification {
  static permission: NotificationPermission = 'default'
  static requestPermission = vi.fn().mockImplementation(async () => {
    MockNotification.permission = 'granted'
    return 'granted'
  })

  constructor(public title: string, public options?: NotificationOptions) {
    mockNotificationInstance(title, options)
  }
}

describe('useNotifications', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    MockNotification.permission = 'default'
    // @ts-expect-error Mocking global Notification
    global.Notification = MockNotification
  })

  it('支持请求通知权限并更新 permission 状态', async () => {
    const { result } = renderHook(() => useNotifications())

    expect(result.current.permission).toBe('default')

    await act(async () => {
      await result.current.requestPermission()
    })

    expect(MockNotification.requestPermission).toHaveBeenCalled()
    expect(result.current.permission).toBe('granted')
  })

  it('当发生 waiting_input 跃迁时自动触发桌面通知', async () => {
    MockNotification.permission = 'granted'
    const { rerender } = renderHook(
      (props: { event: WSStateChangedPayload | null }) =>
        useNotifications({ lastEvent: props.event }),
      { initialProps: { event: null as WSStateChangedPayload | null } }
    )

    expect(mockNotificationInstance).not.toHaveBeenCalled()

    // 状态迁移为 waiting_input
    const waitingEvent: WSStateChangedPayload = {
      pane_id: '%1',
      from_state: 'working',
      to_state: 'waiting_input',
      trigger_rule: 'rule_confirm',
      summary: 'Prompting for approval',
    }

    rerender({ event: waitingEvent })

    expect(mockNotificationInstance).toHaveBeenCalledWith(
      expect.stringContaining('%1'),
      expect.objectContaining({
        body: expect.stringContaining('Prompting for approval'),
      })
    )
  })

  it('当状态跃迁为 working 或 idle 时不触发通知', () => {
    MockNotification.permission = 'granted'
    const { rerender } = renderHook(
      (props: { event: WSStateChangedPayload | null }) =>
        useNotifications({ lastEvent: props.event }),
      { initialProps: { event: null as WSStateChangedPayload | null } }
    )

    const workingEvent: WSStateChangedPayload = {
      pane_id: '%2',
      from_state: 'idle',
      to_state: 'working',
      trigger_rule: 'rule_running',
      summary: 'Running task',
    }

    rerender({ event: workingEvent })

    expect(mockNotificationInstance).not.toHaveBeenCalled()
  })
})
