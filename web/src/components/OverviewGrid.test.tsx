import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { OverviewGrid } from './OverviewGrid'
import type { PaneSummary } from '@/types'

// 来自 docs/spec/info.md 的标准测试数据
const mockPanes: PaneSummary[] = [
  {
    pane_id: '%1',
    session_name: 'tmux-ui-test-session',
    window_id: '@1',
    window_name: 'claude-agent',
    agent_type: 'claude-code',
    state: 'waiting_input',
    current_command: 'node',
    pane_pid: 1001,
    idle_seconds: 15,
    tail_preview: ['[Agent] Refactoring complete.', 'Do you want to proceed? (y/n)'],
    last_active_at: new Date(Date.now() - 15000).toISOString(),
  },
  {
    pane_id: '%2',
    session_name: 'tmux-ui-test-session',
    window_id: '@2',
    window_name: 'codex-runner',
    agent_type: 'codex',
    state: 'working',
    current_command: 'codex',
    pane_pid: 1002,
    idle_seconds: 1,
    tail_preview: ['Generating unit tests... esc to interrupt'],
    last_active_at: new Date(Date.now() - 1000).toISOString(),
  },
  {
    pane_id: '%3',
    session_name: 'tmux-ui-test-session',
    window_id: '@3',
    window_name: 'bash-shell',
    agent_type: 'shell',
    state: 'idle',
    current_command: '/bin/zsh',
    pane_pid: 1003,
    idle_seconds: 120,
    tail_preview: ['user@host:~/project$ '],
    last_active_at: new Date(Date.now() - 120000).toISOString(),
  },
  {
    pane_id: '%4',
    session_name: 'tmux-ui-test-session',
    window_id: '@4',
    window_name: 'aider-task',
    agent_type: 'aider',
    state: 'error',
    current_command: 'python',
    pane_pid: 1004,
    idle_seconds: 45,
    tail_preview: ['fatal: API rate limit exceeded. Exit code 1'],
    last_active_at: new Date(Date.now() - 45000).toISOString(),
  },
  {
    pane_id: '%5',
    session_name: 'tmux-ui-test-session',
    window_id: '@5',
    window_name: 'dead-job',
    agent_type: 'shell',
    state: 'dead',
    current_command: 'zsh',
    pane_pid: 1005,
    idle_seconds: 600,
    tail_preview: ['(pane dead)'],
    last_active_at: new Date(Date.now() - 600000).toISOString(),
  },
]

describe('OverviewGrid & FilterBar', () => {
  it('渲染所有卡片并正确显示统计徽标', () => {
    render(<OverviewGrid panes={mockPanes} onSelectPane={vi.fn()} />)

    // 验证卡片渲染
    expect(screen.getByText('%1')).toBeInTheDocument()
    expect(screen.getByText('%2')).toBeInTheDocument()
    expect(screen.getByText('%3')).toBeInTheDocument()
    expect(screen.getByText('%4')).toBeInTheDocument()
    expect(screen.getByText('%5')).toBeInTheDocument()

    // 验证状态筛选计数字
    expect(screen.getByTestId('count-all')).toHaveTextContent('5')
    expect(screen.getByTestId('count-working')).toHaveTextContent('1')
    expect(screen.getByTestId('count-waiting_input')).toHaveTextContent('1')
    expect(screen.getByTestId('count-idle')).toHaveTextContent('1')
    expect(screen.getByTestId('count-error')).toHaveTextContent('1')
    expect(screen.getByTestId('count-dead')).toHaveTextContent('1')
  })

  it('按状态过滤筛选只显示特定状态卡片', () => {
    render(<OverviewGrid panes={mockPanes} onSelectPane={vi.fn()} />)

    // 点击 waiting_input 筛选按钮
    const waitingTab = screen.getByTestId('filter-waiting_input')
    fireEvent.click(waitingTab)

    expect(screen.getByText('%1')).toBeInTheDocument()
    expect(screen.queryByText('%2')).not.toBeInTheDocument()
    expect(screen.queryByText('%3')).not.toBeInTheDocument()
    expect(screen.queryByText('%4')).not.toBeInTheDocument()
    expect(screen.queryByText('%5')).not.toBeInTheDocument()
  })

  it('使用搜索框按关键词过滤', () => {
    render(<OverviewGrid panes={mockPanes} onSelectPane={vi.fn()} />)

    const searchInput = screen.getByPlaceholderText(/search panes/i)
    fireEvent.change(searchInput, { target: { value: 'aider' } })

    expect(screen.getByText('%4')).toBeInTheDocument()
    expect(screen.queryByText('%1')).not.toBeInTheDocument()
    expect(screen.queryByText('%2')).not.toBeInTheDocument()
  })

  it('点击卡片触发 onSelectPane 回调', () => {
    const handleSelect = vi.fn()
    render(<OverviewGrid panes={mockPanes} onSelectPane={handleSelect} />)

    const card = screen.getByTestId('pane-card-%1')
    fireEvent.click(card)

    expect(handleSelect).toHaveBeenCalledWith('%1')
  })
})

