import { useState, useEffect, useCallback } from 'react'
import type { SnapshotMetadata } from '@/types'
import { apiClient } from '@/api/client'

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
      const data = await apiClient.getSnapshots(paneId)
      if (Array.isArray(data)) {
        setSnapshots(data)
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
      const record = await apiClient.getSnapshotDetail(paneId, snapshotId)
      if (record) {
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
