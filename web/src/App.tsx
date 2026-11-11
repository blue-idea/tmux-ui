import React from 'react'
import { useTmuxWS } from './hooks/useTmuxWS'
import { Terminal, Activity, Wifi, WifiOff } from 'lucide-react'

export const App: React.FC = () => {
  const { status, panes } = useTmuxWS()

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 p-6">
      <header className="flex items-center justify-between pb-6 border-b border-slate-800">
        <div className="flex items-center space-x-3">
          <Terminal className="w-8 h-8 text-blue-500" />
          <h1 className="text-2xl font-bold tracking-tight">tmux-ui</h1>
          <span className="text-xs px-2.5 py-0.5 rounded-full bg-slate-800 text-slate-400 font-mono">
            Multi-Agent Visualizer
          </span>
        </div>

        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2 text-sm text-slate-400">
            <Activity className="w-4 h-4 text-slate-500" />
            <span>Active Panes:</span>
            <span className="font-semibold text-slate-200">{panes.length}</span>
          </div>

          <div
            className={`flex items-center space-x-1.5 px-3 py-1 rounded-full text-xs font-medium ${
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
      </header>

      <main className="mt-8">
        {panes.length === 0 ? (
          <div className="flex flex-col items-center justify-center p-12 text-slate-500 border border-dashed border-slate-800 rounded-xl">
            <Terminal className="w-12 h-12 mb-3 stroke-[1.5]" />
            <p className="text-base font-medium">No active agent panes detected</p>
            <p className="text-sm text-slate-600 mt-1">
              Start a tmux session or run an agent like Claude Code / Aider to begin monitoring
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {panes.map((p) => (
              <div
                key={p.pane_id}
                className="p-4 rounded-xl bg-slate-900 border border-slate-800 hover:border-slate-700 transition"
              >
                <div className="flex items-center justify-between mb-2">
                  <span className="font-mono text-sm font-semibold text-blue-400">{p.pane_id}</span>
                  <span className="text-xs px-2 py-0.5 rounded bg-slate-800 text-slate-300">
                    {p.agent_type || 'shell'}
                  </span>
                </div>
                <div className="text-xs text-slate-400">Session: {p.session_name}</div>
                <div className="text-xs text-slate-400 mt-1">State: {p.state}</div>
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  )
}

export default App
