import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { PlaybackBar } from './PlaybackBar'
import type { SnapshotMetadata } from '@/types'

describe('PlaybackBar Component', () => {
  const mockSnapshots: SnapshotMetadata[] = [
    {
      snapshot_id: 'snap_1',
      pane_id: '%0',
      timestamp: 1731380000000,
      event_type: 'state_transition',
      old_state: 'idle',
      new_state: 'working',
      summary: 'Task Started',
      byte_length: 120,
    },
    {
      snapshot_id: 'snap_2',
      pane_id: '%0',
      timestamp: 1731380030000,
      event_type: 'checkpoint',
      old_state: 'working',
      new_state: 'working',
      summary: 'Periodic Checkpoint',
      byte_length: 240,
    },
    {
      snapshot_id: 'snap_3',
      pane_id: '%0',
      timestamp: 1731380060000,
      event_type: 'state_transition',
      old_state: 'working',
      new_state: 'waiting_input',
      summary: 'User Confirm',
      byte_length: 300,
    },
  ]

  it('展示快照里程碑，并支持点击特定节点回放', () => {
    const onSelect = vi.fn()
    const onReturn = vi.fn()

    render(
      <PlaybackBar
        snapshots={mockSnapshots}
        isLive={true}
        activeSnapshotId={null}
        onSelectSnapshot={onSelect}
        onReturnToLive={onReturn}
      />
    )

    expect(screen.getByText(/3 snapshots/i)).toBeInTheDocument()

    // 点击第二个里程碑节点
    const milestone = screen.getByTestId('milestone-snap_2')
    fireEvent.click(milestone)
    expect(onSelect).toHaveBeenCalledWith('snap_2')
  })

  it('在回放模式下展示警示标识与 Return to Live 按钮', () => {
    const onReturn = vi.fn()

    render(
      <PlaybackBar
        snapshots={mockSnapshots}
        isLive={false}
        activeSnapshotId="snap_2"
        onSelectSnapshot={vi.fn()}
        onReturnToLive={onReturn}
      />
    )

    expect(screen.getByText(/Historical Playback/i)).toBeInTheDocument()

    const returnBtn = screen.getByRole('button', { name: /return to live/i })
    fireEvent.click(returnBtn)
    expect(onReturn).toHaveBeenCalledTimes(1)
  })
})
