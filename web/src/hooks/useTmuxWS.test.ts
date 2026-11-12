import { renderHook, act } from '@testing-library/react'
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { useTmuxWS } from './useTmuxWS'
import { setAuthToken, clearAuthToken } from '../auth/tokenManager'
import type { PaneSummary } from '@/types'

// Mock 浏览器 WebSocket
class MockWebSocket {
  static instances: MockWebSocket[] = []
  url: string
  readyState: number = 0 // 0: CONNECTING, 1: OPEN, 2: CLOSING, 3: CLOSED
  onopen: ((event: unknown) => void) | null = null
  onclose: ((event: unknown) => void) | null = null
  onmessage: ((event: { data: string }) => void) | null = null
  onerror: ((event: unknown) => void) | null = null

  constructor(url: string) {
    this.url = url
    MockWebSocket.instances.push(this)
  }

  send = vi.fn()
  close = vi.fn(() => {
    this.readyState = 3
    if (this.onclose) this.onclose({ code: 1000 })
  })

  // 辅助方法：模拟服务器连接成功
  simulateOpen() {
    this.readyState = 1
    if (this.onopen) this.onopen({})
  }

  // 辅助方法：模拟服务器向客户端推送消息
  simulateMessage(data: unknown) {
    if (this.onmessage) {
      this.onmessage({ data: JSON.stringify(data) })
    }
  }

  // 辅助方法：模拟连接异常断开
  simulateError() {
    if (this.onerror) this.onerror(new Error('Network error'))
    this.readyState = 3
    if (this.onclose) this.onclose({ code: 1006 })
  }
}

describe('useTmuxWS', () => {
  beforeEach(() => {
    MockWebSocket.instances = []
    // @ts-expect-error Mocking global WebSocket
    global.WebSocket = MockWebSocket
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.restoreAllMocks()
    vi.useRealTimers()
  })

  it('连接建立并携带可选 Token', () => {
    const { result } = renderHook(() => useTmuxWS({ token: 'test-secret' }))

    expect(MockWebSocket.instances.length).toBe(1)
    expect(MockWebSocket.instances[0].url).toContain('token=test-secret')
    expect(result.current.status).toBe('connecting')

    act(() => {
      MockWebSocket.instances[0].simulateOpen()
    })

    expect(result.current.status).toBe('connected')
  })

  it('接收 snapshot 帧并初始化 Panes 列表', () => {
    const { result } = renderHook(() => useTmuxWS())
    const ws = MockWebSocket.instances[0]

    act(() => {
      ws.simulateOpen()
    })

    const mockPanes: PaneSummary[] = [
      {
        pane_id: '%1',
        session_name: 'prod',
        window_id: '@1',
        window_name: 'agent',
        agent_type: 'claude-code',
        state: 'working',
        current_command: 'node',
        pane_pid: 1001,
        idle_seconds: 2,
        tail_preview: ['compiling...'],
        last_active_at: new Date().toISOString(),
      },
    ]

    act(() => {
      ws.simulateMessage({
        type: 'snapshot',
        timestamp: new Date().toISOString(),
        payload: { panes: mockPanes },
      })
    })

    expect(result.current.panes.length).toBe(1)
    expect(result.current.panes[0].pane_id).toBe('%1')
    expect(result.current.panes[0].state).toBe('working')
  })

  it('接收 state_changed 增量推送并更新对应 Pane 状态', () => {
    const { result } = renderHook(() => useTmuxWS())
    const ws = MockWebSocket.instances[0]

    act(() => {
      ws.simulateOpen()
    })

    const initialPanes: PaneSummary[] = [
      {
        pane_id: '%1',
        session_name: 'prod',
        window_id: '@1',
        window_name: 'agent',
        agent_type: 'claude-code',
        state: 'working',
        current_command: 'node',
        pane_pid: 1001,
        idle_seconds: 2,
        tail_preview: ['compiling...'],
        last_active_at: new Date().toISOString(),
      },
    ]

    act(() => {
      ws.simulateMessage({
        type: 'snapshot',
        timestamp: new Date().toISOString(),
        payload: { panes: initialPanes },
      })
    })

    // 推送增量状态变迁
    act(() => {
      ws.simulateMessage({
        type: 'state_changed',
        timestamp: new Date().toISOString(),
        payload: {
          pane_id: '%1',
          from_state: 'working',
          to_state: 'waiting_input',
          trigger_rule: 'rule_waiting',
          summary: 'Waiting for confirmation',
        },
      })
    })

    expect(result.current.panes[0].state).toBe('waiting_input')
    expect(result.current.lastEvent?.pane_id).toBe('%1')
    expect(result.current.lastEvent?.to_state).toBe('waiting_input')
  })

  it('断线后自动重连并更新连接状态', () => {
    const { result } = renderHook(() => useTmuxWS())
    const ws = MockWebSocket.instances[0]

    act(() => {
      ws.simulateOpen()
    })
    expect(result.current.status).toBe('connected')

    // 模拟连接异常断开
    act(() => {
      ws.simulateError()
    })
    expect(result.current.status).toBe('disconnected')

    // 快进重连定时器
    act(() => {
      vi.advanceTimersByTime(2000)
    })

    expect(MockWebSocket.instances.length).toBe(2)
  })

  it('v3.REQ-003-AC-001: 当未显式传入 token 时，自动从 tokenManager 提取并追加 ?token=... 到握手 URL', () => {
    setAuthToken('auto_injected_ws_token_999')
    renderHook(() => useTmuxWS())

    expect(MockWebSocket.instances.length).toBe(1)
    expect(MockWebSocket.instances[0].url).toContain('token=auto_injected_ws_token_999')
    clearAuthToken()
  })

  it('当显式传入 options.token 时，优先使用显式指定的 token', () => {
    setAuthToken('cached_token')
    renderHook(() => useTmuxWS({ token: 'explicit_token' }))

    expect(MockWebSocket.instances.length).toBe(1)
    expect(MockWebSocket.instances[0].url).toContain('token=explicit_token')
    expect(MockWebSocket.instances[0].url).not.toContain('cached_token')
    clearAuthToken()
  })
})

