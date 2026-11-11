import React, { useEffect, useRef, useState } from 'react'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import type { PaneDetail } from '@/types'
import { StatusBadge } from './StatusBadge'
import { TimelineDrawer } from './TimelineDrawer'
import {
  ArrowLeft,
  Lock,
  Copy,
  Check,
  Cpu,
  Layers,
  Clock,
  Terminal as TermIcon,
  History,
} from 'lucide-react'

interface TerminalViewerProps {
  detail: PaneDetail
  onBack: () => void
}

export const TerminalViewer: React.FC<TerminalViewerProps> = ({ detail, onBack }) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const termRef = useRef<Terminal | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)
  const [copied, setCopied] = useState(false)
  const [isTimelineOpen, setIsTimelineOpen] = useState(false)

  useEffect(() => {
    if (!containerRef.current) return

    // 初始化 xterm 实例（只读安全保护）
    const term = new Terminal({
      cursorBlink: false,
      disableStdin: true, // 核心安全约束：禁用任何键盘输入
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      fontSize: 13,
      lineHeight: 1.25,
      theme: {
        background: '#090d16',
        foreground: '#e2e8f0',
        cursor: '#64748b',
        selectionBackground: '#334155',
        black: '#0f172a',
        red: '#ef4444',
        green: '#10b981',
        yellow: '#eab308',
        blue: '#3b82f6',
        magenta: '#d946ef',
        cyan: '#06b6d4',
        white: '#f8fafc',
      },
    })

    const fitAddon = new FitAddon()
    term.loadAddon(fitAddon)

    term.open(containerRef.current)
    try {
      fitAddon.fit()
    } catch {
      // 容错处理
    }

    termRef.current = term
    fitAddonRef.current = fitAddon

    // 写入 ANSI 字符流
    if (detail.terminal_output) {
      term.write(detail.terminal_output)
    }

    // 监听容器大小变动
    let resizeObserver: ResizeObserver | null = null
    if (typeof ResizeObserver !== 'undefined') {
      resizeObserver = new ResizeObserver(() => {
        try {
          fitAddon.fit()
        } catch {
          // 忽略微小几何计算错误
        }
      })
      resizeObserver.observe(containerRef.current)
    }

    return () => {
      if (resizeObserver) {
        resizeObserver.disconnect()
      }
      term.dispose()
    }
  }, [detail.terminal_output])

  const handleCopy = async () => {
    if (!detail.terminal_output) return
    try {
      await navigator.clipboard.writeText(detail.terminal_output)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // 忽略剪贴板权限异常
    }
  }

  return (
    <div className="flex flex-col h-[calc(100vh-140px)] gap-4">
      {/* 顶部详情工具栏 */}
      <div className="flex flex-wrap items-center justify-between gap-4 p-4 rounded-xl bg-slate-900 border border-slate-800">
        <div className="flex items-center space-x-4">
          <button
            onClick={onBack}
            className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-medium transition"
          >
            <ArrowLeft className="w-4 h-4" />
            <span>Back</span>
          </button>

          <div className="flex items-center space-x-3">
            <span className="font-mono text-lg font-bold text-blue-400">{detail.pane_id}</span>
            <span className="text-xs text-slate-400 font-mono">
              {detail.session_name}:{detail.window_name}
            </span>
            <StatusBadge state={detail.state} />
          </div>
        </div>

        <div className="flex items-center space-x-3">
          {/* 只读指示条 */}
          <div className="flex items-center space-x-1.5 px-3 py-1 rounded-full bg-slate-800 text-slate-400 text-xs border border-slate-700 font-mono">
            <Lock className="w-3.5 h-3.5 text-amber-500" />
            <span>Read-Only Mode</span>
          </div>

          {/* 时间线抽屉切换 */}
          <button
            onClick={() => setIsTimelineOpen(true)}
            className="flex items-center space-x-1 px-3 py-1 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs transition border border-slate-700/60"
          >
            <History className="w-3.5 h-3.5 text-blue-400" />
            <span>Timeline</span>
            {detail.timeline && detail.timeline.length > 0 && (
              <span className="rounded-full bg-blue-500/20 text-blue-400 text-[10px] px-1.5 py-0.2 font-mono">
                {detail.timeline.length}
              </span>
            )}
          </button>

          {/* 复制输出 */}
          <button
            onClick={handleCopy}
            className="flex items-center space-x-1 px-3 py-1 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs transition"
          >
            {copied ? (
              <>
                <Check className="w-3.5 h-3.5 text-emerald-400" />
                <span className="text-emerald-400">Copied</span>
              </>
            ) : (
              <>
                <Copy className="w-3.5 h-3.5" />
                <span>Copy</span>
              </>
            )}
          </button>
        </div>
      </div>

      {/* 主工作区：终端视口与侧边进程树 */}
      <div className="flex-1 grid grid-cols-1 lg:grid-cols-4 gap-4 min-h-0">
        {/* 左侧 3 列：xterm.js 终端视口 */}
        <div className="lg:col-span-3 flex flex-col rounded-xl border border-slate-800 bg-[#090d16] overflow-hidden shadow-inner">
          <div className="flex items-center justify-between px-4 py-2 border-b border-slate-800/80 bg-slate-900/40 text-xs text-slate-400 font-mono">
            <div className="flex items-center space-x-2">
              <TermIcon className="w-3.5 h-3.5 text-slate-500" />
              <span>ANSI Terminal Stream</span>
            </div>
            <span>PID: {detail.pane_pid}</span>
          </div>

          <div
            data-testid="terminal-container"
            ref={containerRef}
            className="flex-1 p-3 overflow-hidden"
          />
        </div>

        {/* 右侧 1 列：进程树与元数据 */}
        <div className="flex flex-col gap-4 overflow-y-auto">
          {/* 进程列表 */}
          <div className="p-4 rounded-xl bg-slate-900 border border-slate-800">
            <div className="flex items-center space-x-2 text-xs font-semibold text-slate-300 mb-3">
              <Cpu className="w-4 h-4 text-blue-400" />
              <span>Associated Processes</span>
            </div>

            {detail.processes && detail.processes.length > 0 ? (
              <div className="space-y-2">
                {detail.processes.map((proc) => (
                  <div
                    key={proc.pid}
                    className="p-2 rounded bg-slate-950/60 border border-slate-800/80 text-xs font-mono"
                  >
                    <div className="flex items-center justify-between text-slate-500 text-[11px] mb-1">
                      <span>PID: {proc.pid}</span>
                      <span>PPID: {proc.ppid}</span>
                    </div>
                    <div className="text-slate-300 truncate font-mono">{proc.command}</div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-xs text-slate-500 italic">No child processes detected</div>
            )}
          </div>

          {/* 元数据卡片 */}
          <div className="p-4 rounded-xl bg-slate-900 border border-slate-800 space-y-3 text-xs">
            <div className="flex items-center space-x-2 font-semibold text-slate-300">
              <Layers className="w-4 h-4 text-emerald-400" />
              <span>Pane Metadata</span>
            </div>

            <div className="space-y-2 text-slate-400">
              <div className="flex justify-between">
                <span>Agent Type</span>
                <span className="text-slate-200 font-mono">{detail.agent_type || 'shell'}</span>
              </div>
              <div className="flex justify-between">
                <span>Current Command</span>
                <span className="text-slate-200 font-mono">{detail.current_command}</span>
              </div>
              <div className="flex justify-between">
                <span>Window ID</span>
                <span className="text-slate-200 font-mono">{detail.window_id}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="flex items-center space-x-1">
                  <Clock className="w-3 h-3" />
                  <span>Idle Seconds</span>
                </span>
                <span className="text-slate-200 font-mono">{detail.idle_seconds}s</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* 状态变迁时间线抽屉 */}
      <TimelineDrawer
        isOpen={isTimelineOpen}
        onClose={() => setIsTimelineOpen(false)}
        events={detail.timeline || []}
      />
    </div>
  )
}
