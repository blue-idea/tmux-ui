import { renderHook, act, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useHistory } from './useHistory'
import type { SnapshotMetadata, SnapshotRecord } from '@/types'

describe('useHistory Hook', () => {
  const mockSnapshots: SnapshotMetadata[] = [
    {
      snapshot_id: 'snap_1',
      pane_id: '%0',
      timestamp: 1000,
      event_type: 'state_transition',
      old_state: 'idle',
      new_state: 'working',
      summary: 'Started',
      byte_length: 10,
    },
    {
      snapshot_id: 'snap_2',
      pane_id: '%0',
      timestamp: 2000,
      event_type: 'state_transition',
      old_state: 'working',
      new_state: 'waiting_input',
      summary: 'Prompt',
      byte_length: 25,
    },
  ]

  const mockRecord: SnapshotRecord = {
    ...mockSnapshots[1],
    session_name: 'dev',
    window_id: '@0',
    output_buffer: 'Historical ANSI Content',
    tail_lines: ['Historical ANSI Content'],
  }

  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn((url: string) => {
      if (url.includes('/snapshots/snap_2')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ code: 0, data: mockRecord }),
        })
      }
      if (url.includes('/snapshots')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ code: 0, data: mockSnapshots }),
        })
      }
      return Promise.reject(new Error('Unknown url'))
    }))
  })

  it('初始状态为 Live 模式，并自动拉取快照元数据列表', async () => {
    const { result } = renderHook(() => useHistory('%0'))

    expect(result.current.isLive).toBe(true)
    expect(result.current.playbackOutput).toBeNull()

    await waitFor(() => {
      expect(result.current.snapshots).toHaveLength(2)
    })
  })

  it('selectSnapshot 切换至历史回放模式，并加载对应 ANSI 输出', async () => {
    const { result } = renderHook(() => useHistory('%0'))

    await waitFor(() => {
      expect(result.current.snapshots).toHaveLength(2)
    })

    await act(async () => {
      await result.current.selectSnapshot('snap_2')
    })

    expect(result.current.isLive).toBe(false)
    expect(result.current.activeSnapshotId).toBe('snap_2')
    expect(result.current.playbackOutput).toBe('Historical ANSI Content')
  })

  it('returnToLive 一键恢复实时追踪模式', async () => {
    const { result } = renderHook(() => useHistory('%0'))

    await waitFor(() => {
      expect(result.current.snapshots).toHaveLength(2)
    })

    await act(async () => {
      await result.current.selectSnapshot('snap_2')
    })
    expect(result.current.isLive).toBe(false)

    act(() => {
      result.current.returnToLive()
    })

    expect(result.current.isLive).toBe(true)
    expect(result.current.playbackOutput).toBeNull()
    expect(result.current.activeSnapshotId).toBeNull()
  })
})
