import React from 'react'
import type { State } from '@/types'
import { Search } from 'lucide-react'

export type FilterStatus = 'all' | State

interface FilterBarProps {
  selectedStatus: FilterStatus
  onSelectStatus: (status: FilterStatus) => void
  searchQuery: string
  onSearchChange: (query: string) => void
  counts: Record<FilterStatus, number>
}

const TABS: { id: FilterStatus; label: string }[] = [
  { id: 'all', label: 'All' },
  { id: 'working', label: 'Working' },
  { id: 'waiting_input', label: 'Waiting Input' },
  { id: 'idle', label: 'Idle' },
  { id: 'error', label: 'Error' },
  { id: 'dead', label: 'Dead' },
]

export const FilterBar: React.FC<FilterBarProps> = ({
  selectedStatus,
  onSelectStatus,
  searchQuery,
  onSearchChange,
  counts,
}) => {
  return (
    <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between pb-4">
      {/* 状态筛选 Tabs */}
      <div className="flex flex-wrap items-center gap-1.5 p-1 rounded-xl bg-slate-900 border border-slate-800">
        {TABS.map((tab) => {
          const isActive = selectedStatus === tab.id
          const count = counts[tab.id] || 0

          return (
            <button
              key={tab.id}
              data-testid={`filter-${tab.id}`}
              onClick={() => onSelectStatus(tab.id)}
              className={`flex items-center space-x-2 rounded-lg px-3 py-1.5 text-xs font-medium transition-all ${
                isActive
                  ? 'bg-slate-800 text-white shadow-sm'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/50'
              }`}
            >
              <span>{tab.label}</span>
              <span
                data-testid={`count-${tab.id}`}
                className={`rounded-full px-1.5 py-0.2 text-[10px] font-mono ${
                  isActive
                    ? 'bg-slate-700 text-slate-200'
                    : 'bg-slate-800 text-slate-500'
                }`}
              >
                {count}
              </span>
            </button>
          )
        })}
      </div>

      {/* 搜索框 */}
      <div className="relative min-w-[260px]">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          placeholder="Search panes (ID, session, command)..."
          className="w-full rounded-xl bg-slate-900 border border-slate-800 pl-9 pr-4 py-2 text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-blue-500/50 focus:ring-1 focus:ring-blue-500/50 transition"
        />
      </div>
    </div>
  )
}
