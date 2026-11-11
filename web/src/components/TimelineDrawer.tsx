import React from 'react'
import type { TimelineEvent } from '@/types'
import { StatusBadge } from './StatusBadge'
import { BottomSheet } from './BottomSheet'
import { useIsMobile } from '@/hooks/useMediaQuery'
import { X, ArrowRight, History, GitCommit } from 'lucide-react'

interface TimelineDrawerProps {
  isOpen: boolean
  onClose: () => void
  events: TimelineEvent[]
}

function formatTime(isoStr: string): string {
  try {
    const d = new Date(isoStr)
    return d.toLocaleTimeString()
  } catch {
    return isoStr
  }
}

export const TimelineDrawer: React.FC<TimelineDrawerProps> = ({
  isOpen,
  onClose,
  events,
}) => {
  const isMobile = useIsMobile()

  if (!isOpen) return null

  const timelineContent = (
    <div>
      {events.length === 0 ? (
        <div className="flex flex-col items-center justify-center h-48 text-slate-500 text-xs">
          <GitCommit className="w-8 h-8 mb-2 text-slate-600 stroke-[1.5]" />
          <p>No transition events recorded</p>
        </div>
      ) : (
        <div className="relative border-l border-slate-800 ml-3 space-y-6">
          {events.map((evt, idx) => (
            <div key={evt.id || idx} className="relative pl-6">
              {/* 时间线圆点 */}
              <div className="absolute -left-1.5 top-1 w-3 h-3 rounded-full bg-slate-900 border-2 border-blue-500" />

              <div className="space-y-2 rounded-xl bg-slate-950/60 border border-slate-800/80 p-3.5 text-xs">
                <div className="flex items-center justify-between text-[11px] text-slate-500 font-mono">
                  <span>{formatTime(evt.timestamp)}</span>
                  <span className="text-slate-400">rule: {evt.trigger_rule}</span>
                </div>

                {/* 状态迁移动态 */}
                <div className="flex items-center space-x-2 my-1">
                  <StatusBadge state={evt.from_state} size="sm" />
                  <ArrowRight className="w-3.5 h-3.5 text-slate-500" />
                  <StatusBadge state={evt.to_state} size="sm" />
                </div>

                {/* 变迁说明 */}
                {evt.summary && (
                  <p className="text-slate-300 font-medium leading-relaxed mt-2 pt-2 border-t border-slate-800/60">
                    {evt.summary}
                  </p>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )

  // 移动端视口下自适应为 Bottom Sheet
  if (isMobile) {
    return (
      <BottomSheet isOpen={isOpen} onClose={onClose} title="State Transition History">
        {timelineContent}
      </BottomSheet>
    )
  }

  // 桌面端保持右侧滑出抽屉
  return (
    <div className="fixed inset-0 z-50 overflow-hidden bg-black/60 backdrop-blur-sm flex justify-end transition-opacity">
      <div className="w-full max-w-md bg-slate-900 border-l border-slate-800 h-full flex flex-col shadow-2xl animate-in slide-in-from-right duration-200">
        {/* 顶部标题与关闭 */}
        <div className="p-4 border-b border-slate-800 flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <History className="w-5 h-5 text-blue-400" />
            <h2 className="text-sm font-semibold text-slate-100">State Transition History</h2>
          </div>
          <button
            onClick={onClose}
            aria-label="Close"
            className="flex items-center justify-center w-10 h-10 rounded-lg hover:bg-slate-800 text-slate-400 hover:text-slate-200 transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* 时间线内容列表 */}
        <div className="flex-1 overflow-y-auto p-4">
          {timelineContent}
        </div>
      </div>
    </div>
  )
}
