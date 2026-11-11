import React, { useState, useEffect } from 'react'
import { useTmuxWS } from './hooks/useTmuxWS'
import { OverviewGrid } from './components/OverviewGrid'
import { TerminalViewer } from './components/TerminalViewer'
import type { PaneDetail } from './types'
import { Terminal, Activity, Wifi, WifiOff } from 'lucide-react'

export const App: React.FC = () => {
  const { status, panes } = useTmuxWS()
  const [selectedPaneId, setSelectedPaneId] = useState<string | null>(null)
  const [paneDetail, setPaneDetail] = useState<PaneDetail | null>(null)

  // 当选中某个 Pane 时，异步请求其全量详情
  useEffect(() => {
    if (!selectedPaneId) {
      setPaneDetail(null)
      return
    }

    const matched = panes.find((p) => p.pane_id === selectedPaneId)

    // 首先基于本地 summary 提供即时响应
    if (matched) {
      setPaneDetail({
        ...matched,
        terminal_output: matched.tail_preview?.join('\n') || '',
        timeline: [],
        processes: [
          {
            pid: matched.pane_pid,
            ppid: 1,
            command: matched.current_command,
          },
        ],
      })
    }

    // 然后异步向后端 REST API 请求全量输出
    const encodedId = encodeURIComponent(selectedPaneId)
    fetch(`/api/v1/panes/${encodedId}`)
      .then((res) => {
        if (!res.ok) throw new Error('Failed to fetch detail')
        return res.json()
      })
      .then((resJson) => {
        if (resJson && resJson.data) {
          setPaneDetail(resJson.data)
        }
      })
      .catch(() => {
        // 使用本地合成的兜底数据
      })
  }, [selectedPaneId, panes])

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col">
      {/* 顶部全局导航 */}
      <header className="sticky top-0 z-30 border-b border-slate-800/80 bg-slate-950/80 backdrop-blur-md px-6 py-4">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="p-2 rounded-lg bg-blue-500/10 text-blue-400 border border-blue-500/20">
              <Terminal className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <h1 className="text-lg font-bold tracking-tight text-white">tmux-ui</h1>
                <span className="text-[10px] px-2 py-0.5 rounded-full bg-slate-800 text-slate-400 font-mono">
                  v1.0
                </span>
              </div>
              <p className="text-xs text-slate-500">Autonomous Agent Monitoring Console</p>
            </div>
          </div>

          <div className="flex items-center space-x-4">
            <div className="flex items-center space-x-2 text-xs text-slate-400 bg-slate-900 border border-slate-800 px-3 py-1.5 rounded-lg">
              <Activity className="w-3.5 h-3.5 text-blue-400" />
              <span>Panes:</span>
              <span className="font-semibold text-slate-200">{panes.length}</span>
            </div>

            <div
              className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-lg text-xs font-medium ${
                status === 'connected'
                  ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                  : status === 'connecting'
                  ? 'bg-yellow-500/10 text-yellow-400 border border-yellow-500/20'
                  : 'bg-rose-500/10 text-rose-400 border border-rose-500/20'
              }`}
            >
              {status === 'connected' ? (
                <Wifi className="w-3.5 h-3.5" />
              ) : (
                <WifiOff className="w-3.5 h-3.5" />
              )}
              <span className="capitalize">{status}</span>
            </div>
          </div>
        </div>
      </header>

      {/* 主视图内容 */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-6 py-6">
        {selectedPaneId && paneDetail ? (
          <TerminalViewer detail={paneDetail} onBack={() => setSelectedPaneId(null)} />
        ) : (
          <OverviewGrid panes={panes} onSelectPane={(id) => setSelectedPaneId(id)} />
        )}
      </main>

      {/* 底部状态条 */}
      <footer className="border-t border-slate-800/60 px-6 py-3 text-xs text-slate-600 flex items-center justify-between max-w-7xl w-full mx-auto">
        <span>Selected Pane: {selectedPaneId || 'None (Overview Mode)'}</span>
        <span>tmux-ui · Spec-Driven Development v1</span>
      </footer>
    </div>
  )
}

export default App
