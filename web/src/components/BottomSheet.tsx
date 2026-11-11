import React from 'react'
import { X } from 'lucide-react'

interface BottomSheetProps {
  isOpen: boolean
  onClose: () => void
  title: string
  children: React.ReactNode
}

export const BottomSheet: React.FC<BottomSheetProps> = ({
  isOpen,
  onClose,
  title,
  children,
}) => {
  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex flex-col justify-end">
      {/* 半透明背景遮罩 */}
      <div
        data-testid="bottom-sheet-backdrop"
        onClick={onClose}
        className="fixed inset-0 bg-black/70 backdrop-blur-sm transition-opacity animate-in fade-in duration-200"
      />

      {/* 自底向上滑出的面板容器 */}
      <div className="relative z-10 w-full max-h-[85vh] rounded-t-3xl bg-slate-900 border-t border-slate-800 flex flex-col shadow-2xl animate-in slide-in-from-bottom duration-200">
        {/* 顶部手势拖拽把手指示条 */}
        <div className="pt-3 pb-1 flex justify-center">
          <div className="w-12 h-1.5 rounded-full bg-slate-700/80" />
        </div>

        {/* 标题栏与触控关闭按钮（>= 44x44px） */}
        <div className="px-5 py-3 border-b border-slate-800/80 flex items-center justify-between">
          <h2 className="text-sm font-semibold text-slate-100">{title}</h2>
          <button
            onClick={onClose}
            aria-label="Close"
            className="flex items-center justify-center w-11 h-11 -mr-2 rounded-xl text-slate-400 hover:text-slate-200 hover:bg-slate-800/60 active:bg-slate-800 transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* 可滚动内容区 */}
        <div className="flex-1 overflow-y-auto overscroll-contain p-4 pb-8">
          {children}
        </div>
      </div>
    </div>
  )
}
