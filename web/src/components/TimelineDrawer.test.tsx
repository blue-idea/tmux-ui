import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { TimelineDrawer } from './TimelineDrawer'
import type { TimelineEvent } from '@/types'

const mockEvents: TimelineEvent[] = [
  {
    id: 'evt-1',
    timestamp: '2026-11-11T12:00:00Z',
    pane_id: '%1',
    from_state: 'idle',
    to_state: 'working',
    trigger_rule: 'rule_working',
    summary: 'Started compilation process',
  },
  {
    id: 'evt-2',
    timestamp: '2026-11-11T12:05:00Z',
    pane_id: '%1',
    from_state: 'working',
    to_state: 'waiting_input',
    trigger_rule: 'rule_waiting',
    summary: 'Agent waiting for confirmation',
  },
]

describe('TimelineDrawer', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('当 isOpen 为 true 时展示事件节点与迁移详情', () => {
    render(<TimelineDrawer isOpen={true} onClose={vi.fn()} events={mockEvents} />)

    expect(screen.getByText(/state transition history/i)).toBeInTheDocument()
    expect(screen.getByText('Started compilation process')).toBeInTheDocument()
    expect(screen.getByText('Agent waiting for confirmation')).toBeInTheDocument()
    expect(screen.getByText(/rule_waiting/)).toBeInTheDocument()
  })

  it('点击关闭按钮触发 onClose 回调', () => {
    const handleClose = vi.fn()
    render(<TimelineDrawer isOpen={true} onClose={handleClose} events={mockEvents} />)

    const closeBtn = screen.getByRole('button', { name: /close/i })
    fireEvent.click(closeBtn)

    expect(handleClose).toHaveBeenCalled()
  })

  it('当事件为空时显示空提示', () => {
    render(<TimelineDrawer isOpen={true} onClose={vi.fn()} events={[]} />)

    expect(screen.getByText(/no transition events recorded/i)).toBeInTheDocument()
  })
})
