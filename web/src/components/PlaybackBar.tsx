import React from 'react'
import type { SnapshotMetadata } from '@/types'
import { Radio, RotateCcw, Clock } from 'lucide-react'

interface PlaybackBarProps {
  snapshots: SnapshotMetadata[]
  isLive: boolean
  activeSnapshotId: string | null
  onSelectSnapshot: (id: string) => void
  onReturnToLive: () => void
}

function formatTime(timestamp: number): string {
  try {
    return new Date(timestamp).toLocaleTimeString()
  } catch {
    return String(timestamp)
  }
}

export const PlaybackBar: React.FC<PlaybackBarProps> = ({
  snapshots,
  isLive,
  activeSnapshotId,
  onSelectSnapshot,
  onReturnToLive,
}) => {
  if (!snapshots || snapshots.length === 0) {
    return null
  }

  const activeIndex = snapshots.findIndex((s) => s.snapshot_id === activeSnapshotId)
  const currentSnapshot = activeIndex >= 0 ? snapshots[activeIndex] : null

  return (
    <div className="flex flex-col gap-2 rounded-xl border border-slate-800 bg-slate-900/95 p-3.5 shadow-lg backdrop-blur-md">
      {/* 顶部状态与指示条 */}
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center space-x-2">
          {isLive ? (
            <div className="flex items-center space-x-1.5 px-2.5 py-1 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 text-xs font-mono font-medium">
              <Radio className="w-3.5 h-3.5 animate-pulse" />
              <span>LIVE TRACKING</span>
            </div>
          ) : (
            <div className="flex items-center space-x-2 px-3 py-1 rounded-full bg-amber-500/10 text-amber-400 border border-amber-500/30 text-xs font-mono font-semibold">
              <Clock className="w-3.5 h-3.5 shrink-0" />
              <span>Historical Playback ({currentSnapshot ? formatTime(currentSnapshot.timestamp) : ''})</span>
            </div>
          )}

          <span className="text-xs text-slate-500 font-mono">
            {snapshots.length} snapshot{snapshots.length > 1 ? 's' : ''}
          </span>
        </div>

        {/* 恢复 Live 按钮 */}
        {!isLive && (
          <button
            onClick={onReturnToLive}
            aria-label="Return to Live"
            className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-slate-950 text-xs font-bold shadow-md hover:shadow-emerald-500/20 active:scale-95 transition min-h-[36px]"
          >
            <RotateCcw className="w-3.5 h-3.5" />
            <span>Return to Live</span>
          </button>
        )}
      </div>

      {/* 时间轴里程碑点与滑动条 */}
      <div className="relative pt-2 pb-1">
        {/* 滑动轨道 */}
        <div className="relative flex items-center h-4">
          <div className="absolute inset-x-0 h-1.5 rounded-full bg-slate-800" />
          <div
            className="absolute left-0 h-1.5 rounded-full bg-blue-500/60 transition-all"
            style={{
              width: `${
                isLive
                  ? 100
                  : snapshots.length > 1 && activeIndex >= 0
                    ? (activeIndex / (snapshots.length - 1)) * 100
                    : 0
              }%`,
            }}
          />

          {/* 快照节点圆点 (Milestones) */}
          <div className="relative flex justify-between w-full z-10 px-0.5">
            {snapshots.map((snap) => {
              const isSelected = !isLive && snap.snapshot_id === activeSnapshotId
              const stateColor =
                snap.new_state === 'waiting_input'
                  ? 'bg-yellow-400'
                  : snap.new_state === 'error'
                    ? 'bg-rose-500'
                    : snap.new_state === 'working'
                      ? 'bg-blue-400'
                      : 'bg-slate-400'

              return (
                <button
                  key={snap.snapshot_id}
                  data-testid={`milestone-${snap.snapshot_id}`}
                  onClick={() => onSelectSnapshot(snap.snapshot_id)}
                  title={`[${formatTime(snap.timestamp)}] ${snap.summary || snap.new_state}`}
                  className={`group relative flex items-center justify-center p-1 -m-1 focus:outline-none`}
                >
                  <div
                    className={`w-3 h-3 rounded-full border-2 border-slate-900 transition-all ${stateColor} ${
                      isSelected
                        ? 'ring-2 ring-white scale-125'
                        : 'hover:scale-125 opacity-80 hover:opacity-100'
                    }`}
                  />
                  {/* Tooltip */}
                  <span className="absolute -top-7 hidden group-hover:flex px-2 py-0.5 rounded bg-slate-800 text-[10px] text-slate-200 font-mono whitespace-nowrap shadow border border-slate-700 pointer-events-none">
                    {formatTime(snap.timestamp)} · {snap.new_state}
                  </span>
                </button>
              )
            })}
          </div>
        </div>
      </div>
    </div>
  )
}
