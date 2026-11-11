import { useState, useEffect, useCallback } from 'react'
import type { SnapshotMetadata, SnapshotRecord } from '@/types'

export interface UseHistoryReturn {
  snapshots: SnapshotMetadata[]
  isLive: boolean
  activeSnapshotId: string | null
  playbackOutput: string | null
  isLoading: boolean
  selectSnapshot: (snapshotId: string) => Promise<void>
  returnToLive: () => void
  refreshSnapshots: () => Promise<void>
}

/**
 * useHistory 负责管理终端历史快照拉取、时间轴点选以及 Live/回放模式切换
 */
export function useHistory(paneId: string): UseHistoryReturn {
  const [snapshots, setSnapshots] = useState<SnapshotMetadata[]>([])
  const [isLive, setIsLive] = useState(true)
  const [activeSnapshotId, setActiveSnapshotId] = useState<string | null>(null)
  const [playbackOutput, setPlaybackOutput] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  // 异步获取快照元数据列表
  const refreshSnapshots = useCallback(async () => {
    if (!paneId) return
    try {
      const encodedId = encodeURIComponent(paneId)
      const res = await fetch(`/api/v1/panes/${encodedId}/snapshots`)
      if (!res.ok) return
      const json = await res.json()
      if (json && Array.isArray(json.data)) {
        setSnapshots(json.data)
      }
    } catch {
      // 忽略请求异常
    }
  }, [paneId])

  useEffect(() => {
    refreshSnapshots()
  }, [refreshSnapshots])

  // 点选特定历史快照
  const selectSnapshot = useCallback(async (snapshotId: string) => {
    if (!paneId || !snapshotId) return
    setIsLoading(true)
    try {
      const encodedId = encodeURIComponent(paneId)
      const res = await fetch(`/api/v1/panes/${encodedId}/snapshots/${snapshotId}`)
      if (!res.ok) throw new Error('Failed to fetch snapshot detail')
      const json = await res.json()
      if (json && json.data) {
        const record = json.data as SnapshotRecord
        setPlaybackOutput(record.output_buffer || '')
        setActiveSnapshotId(snapshotId)
        setIsLive(false)
      }
    } catch {
      // 忽略异常
    } finally {
      setIsLoading(false)
    }
  }, [paneId])

  // 恢复至实时追踪模式
  const returnToLive = useCallback(() => {
    setIsLive(true)
    setActiveSnapshotId(null)
    setPlaybackOutput(null)
  }, [])

  return {
    snapshots,
    isLive,
    activeSnapshotId,
    playbackOutput,
    isLoading,
    selectSnapshot,
    returnToLive,
    refreshSnapshots,
  }
}
