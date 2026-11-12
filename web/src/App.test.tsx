import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import App from './App'
import type { PaneSummary, PaneDetail } from './types'

// 模拟测试数据（通过 vi.hoisted 规避 hoisting 作用域限制）
const { mockPanes, mockDetail } = vi.hoisted(() => {
  const panes: PaneSummary[] = [
    {
      pane_id: '%1',
      session_name: 'test-session',
      window_id: '@1',
      window_name: 'claude-runner',
      agent_type: 'claude-code',
      state: 'working',
      current_command: 'node',
      pane_pid: 2001,
      idle_seconds: 3,
      tail_preview: ['Refactoring code...', 'Thinking...'],
      last_active_at: new Date().toISOString(),
    },
    {
      pane_id: '%2',
      session_name: 'test-session',
      window_id: '@2',
      window_name: 'aider-runner',
      agent_type: 'aider',
      state: 'waiting_input',
      current_command: 'python',
      pane_pid: 2002,
      idle_seconds: 12,
      tail_preview: ['Do you want to proceed? (y/n)'],
      last_active_at: new Date().toISOString(),
    },
  ]

  const detail: PaneDetail = {
    ...panes[0],
    terminal_output: '\x1b[32m[Claude] Running...\x1b[0m\nThinking...',
    timeline: [],
    processes: [{ pid: 2001, ppid: 1, command: 'node' }],
  }

  return { mockPanes: panes, mockDetail: detail }
})

// Mock xterm.js 与 FitAddon
vi.mock('@xterm/xterm', () => {
  return {
    Terminal: vi.fn().mockImplementation(() => ({
      open: vi.fn(),
      write: vi.fn(),
      dispose: vi.fn(),
      loadAddon: vi.fn(),
      onData: vi.fn(),
      clear: vi.fn(),
      reset: vi.fn(),
      scrollToBottom: vi.fn(),
      options: {},
    })),
  }
})

vi.mock('@xterm/addon-fit', () => {
  return {
    FitAddon: vi.fn().mockImplementation(() => ({
      fit: vi.fn(),
    })),
  }
})

// 模拟 useTmuxWS 返回活动 Pane
vi.mock('./hooks/useTmuxWS', () => ({
  useTmuxWS: () => ({
    status: 'connected',
    panes: mockPanes,
    lastEvent: null,
    reconnect: vi.fn(),
  }),
}))

// 模拟 API 客户端
vi.mock('./api/client', () => ({
  apiClient: {
    getPaneDetail: vi.fn().mockResolvedValue(mockDetail),
    getSnapshots: vi.fn().mockResolvedValue([]),
    getSnapshotDetail: vi.fn().mockResolvedValue(null),
  },
  addAuthErrorListener: vi.fn().mockReturnValue(() => {}),
}))

describe('App 集成流程测试: 点击 Agent 卡片查看对应 tmux 终端', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('初始状态处于 Overview 总览墙，点击特定 Agent 卡片切换进入 TerminalViewer 终端视图，点击 Back 返回总览', async () => {
    render(<App />)

    // 1. 验证初始状态处于总览网格，渲染了所有 Agent 卡片
    expect(screen.getByText('tmux-ui')).toBeInTheDocument()
    expect(screen.getByTestId('pane-card-%1')).toBeInTheDocument()
    expect(screen.getByTestId('pane-card-%2')).toBeInTheDocument()
    expect(screen.getByText('claude-code')).toBeInTheDocument()

    // 2. 点击 Agent 卡片 %1
    const card = screen.getByTestId('pane-card-%1')
    fireEvent.click(card)

    // 3. 验证视图切换至 TerminalViewer 终端视图
    await waitFor(() => {
      expect(screen.getByTestId('terminal-container')).toBeInTheDocument()
    })
    expect(screen.getByText(/read-only/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /back/i })).toBeInTheDocument()

    // 4. 点击 Back 按钮返回总览卡片墙
    const backBtn = screen.getByRole('button', { name: /back/i })
    fireEvent.click(backBtn)

    // 5. 验证成功回到总览网格
    expect(screen.getByTestId('pane-card-%1')).toBeInTheDocument()
    expect(screen.queryByTestId('terminal-container')).not.toBeInTheDocument()
  })
})
