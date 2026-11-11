import React from 'react'
import type { PaneSummary } from '@/types'
import { StatusBadge } from './StatusBadge'
import { Terminal, Clock, Cpu } from 'lucide-react'

interface PaneCardProps {
  pane: PaneSummary
  onClick?: (paneId: string) => void
}

function formatIdleTime(seconds: number): string {
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`
  return `${Math.floor(seconds / 3600)}h`
}

export const PaneCard: React.FC<PaneCardProps> = ({ pane, onClick }) => {
  return (
    <div
      data-testid={`pane-card-${pane.pane_id}`}
      onClick={() => onClick?.(pane.pane_id)}
      className="group relative flex flex-col justify-between rounded-xl border border-slate-800 bg-slate-900/80 p-4 sm:p-5 shadow-sm transition-all duration-200 hover:border-slate-700 hover:bg-slate-900 hover:shadow-md cursor-pointer active:scale-[0.99]"
    >
      <div>
        {/* 顶部标题与 Agent Badge */}
        <div className="flex items-start justify-between gap-2">
          <div className="flex items-center space-x-2 truncate">
            <span className="font-mono text-base font-semibold text-blue-400 group-hover:text-blue-300">
              {pane.pane_id}
            </span>
            <span className="text-xs text-slate-500 font-mono truncate max-w-[150px] sm:max-w-none">
              {pane.session_name}:{pane.window_name}
            </span>
          </div>

          <StatusBadge state={pane.state} />
        </div>

        {/* 智能体类型与进程信息 - 移动端自适应折行 */}
        <div className="mt-3 flex flex-wrap items-center gap-2 text-xs text-slate-400">
          <span className="inline-flex items-center space-x-1 rounded bg-slate-800/80 px-2 py-0.5 text-slate-300 font-mono">
            <Cpu className="w-3 h-3 text-slate-400 shrink-0" />
            <span className="truncate max-w-[90px]">{pane.agent_type || 'shell'}</span>
          </span>

          <span className="flex items-center space-x-1 font-mono text-slate-400">
            <Terminal className="w-3 h-3 text-slate-500 shrink-0" />
            <span className="truncate max-w-[120px]">{pane.current_command || 'zsh'}</span>
          </span>

          <span className="flex items-center space-x-1 text-slate-500">
            <Clock className="w-3 h-3 text-slate-500 shrink-0" />
            <span>idle {formatIdleTime(pane.idle_seconds)}</span>
          </span>
        </div>

        {/* 终端末尾预览输出 */}
        <div className="mt-3 sm:mt-4 rounded-lg bg-black/60 p-3 font-mono text-xs text-slate-300 border border-slate-800/60 overflow-hidden">
          {pane.tail_preview && pane.tail_preview.length > 0 ? (
            pane.tail_preview.slice(-3).map((line, idx) => (
              <div key={idx} className="truncate whitespace-pre leading-relaxed font-mono">
                {line || ' '}
              </div>
            ))
          ) : (
            <span className="text-slate-600 italic">No output preview</span>
          )}
        </div>
      </div>
    </div>
  )
}
