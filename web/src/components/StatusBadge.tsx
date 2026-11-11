import React from 'react'
import type { State } from '@/types'
import { Loader2, AlertCircle, Clock, CheckCircle2, XCircle } from 'lucide-react'

interface StatusBadgeProps {
  state: State
  size?: 'sm' | 'md'
}

export const StatusBadge: React.FC<StatusBadgeProps> = ({ state, size = 'sm' }) => {
  const configs: Record<
    State,
    { label: string; bg: string; text: string; border: string; icon: React.ReactNode }
  > = {
    working: {
      label: 'Working',
      bg: 'bg-blue-500/10',
      text: 'text-blue-400',
      border: 'border-blue-500/30',
      icon: <Loader2 className="w-3.5 h-3.5 animate-spin" />,
    },
    waiting_input: {
      label: 'Waiting Input',
      bg: 'bg-yellow-500/10',
      text: 'text-yellow-400',
      border: 'border-yellow-500/30',
      icon: <Clock className="w-3.5 h-3.5 animate-pulse" />,
    },
    idle: {
      label: 'Idle',
      bg: 'bg-emerald-500/10',
      text: 'text-emerald-400',
      border: 'border-emerald-500/30',
      icon: <CheckCircle2 className="w-3.5 h-3.5" />,
    },
    error: {
      label: 'Error',
      bg: 'bg-rose-500/10',
      text: 'text-rose-400',
      border: 'border-rose-500/30',
      icon: <AlertCircle className="w-3.5 h-3.5 animate-bounce" />,
    },
    dead: {
      label: 'Dead',
      bg: 'bg-slate-700/30',
      text: 'text-slate-400',
      border: 'border-slate-700',
      icon: <XCircle className="w-3.5 h-3.5" />,
    },
  }

  const current = configs[state] || configs.idle

  return (
    <span
      className={`inline-flex items-center space-x-1.5 rounded-full border font-medium ${
        size === 'sm' ? 'px-2.5 py-0.5 text-xs' : 'px-3 py-1 text-sm'
      } ${current.bg} ${current.text} ${current.border}`}
    >
      {current.icon}
      <span>{current.label}</span>
    </span>
  )
}

