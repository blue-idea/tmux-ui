import React, { useState, useMemo } from 'react'
import type { PaneSummary } from '@/types'
import { FilterBar, type FilterStatus } from './FilterBar'
import { PaneCard } from './PaneCard'
import { Terminal, Inbox } from 'lucide-react'

interface OverviewGridProps {
  panes: PaneSummary[]
  onSelectPane?: (paneId: string) => void
}

export const OverviewGrid: React.FC<OverviewGridProps> = ({ panes, onSelectPane }) => {
  const [selectedStatus, setSelectedStatus] = useState<FilterStatus>('all')
  const [searchQuery, setSearchQuery] = useState('')

  // 统计各状态数量
  const counts = useMemo(() => {
    const summary: Record<FilterStatus, number> = {
      all: panes.length,
      working: 0,
      waiting_input: 0,
      idle: 0,
      error: 0,
      dead: 0,
    }

    for (const p of panes) {
      if (summary[p.state] !== undefined) {
        summary[p.state]++
      }
    }
    return summary
  }, [panes])

  // 按状态与关键词过滤
  const filteredPanes = useMemo(() => {
    const query = searchQuery.toLowerCase().trim()

    return panes.filter((p) => {
      // 状态筛选
      if (selectedStatus !== 'all' && p.state !== selectedStatus) {
        return false
      }

      // 关键词搜索
      if (query) {
        const matchId = p.pane_id.toLowerCase().includes(query)
        const matchSession = p.session_name.toLowerCase().includes(query)
        const matchWindow = p.window_name.toLowerCase().includes(query)
        const matchAgent = (p.agent_type || '').toLowerCase().includes(query)
        const matchCommand = (p.current_command || '').toLowerCase().includes(query)

        if (!matchId && !matchSession && !matchWindow && !matchAgent && !matchCommand) {
          return false
        }
      }

      return true
    })
  }, [panes, selectedStatus, searchQuery])

  return (
    <div className="space-y-6">
      <FilterBar
        selectedStatus={selectedStatus}
        onSelectStatus={setSelectedStatus}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        counts={counts}
      />

      {panes.length === 0 ? (
        <div className="flex flex-col items-center justify-center p-16 text-slate-500 border border-dashed border-slate-800 rounded-2xl bg-slate-900/30">
          <Terminal className="w-12 h-12 mb-3 stroke-[1.5] text-slate-600" />
          <p className="text-base font-medium text-slate-300">No active agent panes detected</p>
          <p className="text-sm text-slate-500 mt-1 max-w-sm text-center">
            Start a tmux session or run an agent like Claude Code / Aider to begin monitoring
          </p>
        </div>
      ) : filteredPanes.length === 0 ? (
        <div className="flex flex-col items-center justify-center p-16 text-slate-500 border border-dashed border-slate-800/80 rounded-2xl bg-slate-900/20">
          <Inbox className="w-10 h-10 mb-2 stroke-[1.5] text-slate-600" />
          <p className="text-sm font-medium text-slate-400">No matching panes found</p>
          <p className="text-xs text-slate-600 mt-1">Try adjusting your filter or search keywords</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {filteredPanes.map((pane) => (
            <PaneCard key={pane.pane_id} pane={pane} onClick={onSelectPane} />
          ))}
        </div>
      )}
    </div>
  )
}

