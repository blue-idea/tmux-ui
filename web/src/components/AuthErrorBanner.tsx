import React from 'react'
import { ShieldAlert, X } from 'lucide-react'

export interface AuthErrorBannerProps {
  message: string | null
  onDismiss: () => void
}

/**
 * AuthErrorBanner 鉴权失败与 Token 无效顶部告警横幅
 */
export const AuthErrorBanner: React.FC<AuthErrorBannerProps> = ({ message, onDismiss }) => {
  if (!message) return null

  return (
    <div
      role="alert"
      className="bg-rose-500/15 border-b border-rose-500/30 text-rose-200 px-4 py-3 sm:px-6 transition-all duration-200"
    >
      <div className="max-w-7xl mx-auto flex items-center justify-between gap-3">
        <div className="flex items-center space-x-2.5 sm:space-x-3 overflow-hidden">
          <div className="p-1 rounded-md bg-rose-500/20 text-rose-400 shrink-0">
            <ShieldAlert className="w-4 h-4 sm:w-5 sm:h-5" />
          </div>
          <div className="text-xs sm:text-sm">
            <span className="font-semibold text-rose-100 mr-2">Authentication Failed</span>
            <span className="text-rose-300 break-words">{message}</span>
          </div>
        </div>

        <button
          onClick={onDismiss}
          aria-label="Dismiss"
          className="p-1.5 rounded-lg text-rose-400 hover:text-white hover:bg-rose-500/20 transition shrink-0 min-h-[32px] min-w-[32px] flex items-center justify-center"
        >
          <X className="w-4 h-4" />
        </button>
      </div>
    </div>
  )
}
