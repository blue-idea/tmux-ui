import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { TerminalViewer } from './TerminalViewer'
import type { PaneDetail } from '@/types'

// Mock @xterm/xterm 与 @xterm/addon-fit
const mockWrite = vi.fn()
const mockDispose = vi.fn()
const mockFit = vi.fn()

vi.mock('@xterm/xterm', () => {
  return {
    Terminal: vi.fn().mockImplementation(() => ({
      open: vi.fn(),
      write: mockWrite,
      dispose: mockDispose,
      loadAddon: vi.fn(),
      onData: vi.fn(),
      clear: vi.fn(),
      options: {},
    })),
  }
})

vi.mock('@xterm/addon-fit', () => {
  return {
    FitAddon: vi.fn().mockImplementation(() => ({
      fit: mockFit,
    })),
  }
})

const mockDetail: PaneDetail = {
  pane_id: '%1',
  session_name: 'tmux-ui-test-session',
  window_id: '@1',
  window_name: 'claude-agent',
  agent_type: 'claude-code',
  state: 'waiting_input',
  current_command: 'node',
  pane_pid: 1001,
  idle_seconds: 15,
  tail_preview: ['[Agent] Refactoring complete.'],
  last_active_at: new Date().toISOString(),
  terminal_output: '\x1b[32m[Agent] Success\x1b[0m\nDo you want to proceed? (y/n)',
  timeline: [],
  processes: [
    { pid: 1001, ppid: 1000, command: 'node /usr/local/bin/claude' },
    { pid: 1002, ppid: 1001, command: 'git status' },
  ],
}

describe('TerminalViewer', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('挂载 xterm 渲染容器并写入 ANSI 字符流', () => {
    render(<TerminalViewer detail={mockDetail} onBack={vi.fn()} />)

    expect(screen.getByTestId('terminal-container')).toBeInTheDocument()
    expect(mockWrite).toHaveBeenCalledWith(mockDetail.terminal_output)
  })

  it('展示只读模式标记与防篡改提示', () => {
    render(<TerminalViewer detail={mockDetail} onBack={vi.fn()} />)

    expect(screen.getByText(/read-only/i)).toBeInTheDocument()
  })

  it('展示关联的进程列表信息', () => {
    render(<TerminalViewer detail={mockDetail} onBack={vi.fn()} />)

    expect(screen.getAllByText(/PID: 1001/).length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('node /usr/local/bin/claude')).toBeInTheDocument()
    expect(screen.getAllByText(/PID: 1002/).length).toBeGreaterThanOrEqual(1)
  })
})
