import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { AuthErrorBanner } from './AuthErrorBanner'

describe('AuthErrorBanner 组件', () => {
  it('当提供错误信息时，渲染标准英文告警文案与错误图标', () => {
    render(<AuthErrorBanner message="Authentication token is missing or invalid" onDismiss={() => {}} />)

    expect(screen.getByText('Authentication Failed')).toBeInTheDocument()
    expect(screen.getByText(/Authentication token is missing or invalid/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /dismiss/i })).toBeInTheDocument()
  })

  it('点击 Dismiss 按钮触发 onDismiss 回调', () => {
    const onDismiss = vi.fn()
    render(<AuthErrorBanner message="Unauthorized request" onDismiss={onDismiss} />)

    const dismissBtn = screen.getByRole('button', { name: /dismiss/i })
    fireEvent.click(dismissBtn)
    expect(onDismiss).toHaveBeenCalledTimes(1)
  })

  it('当 message 为空时，不渲染任何 DOM 内容', () => {
    const { container } = render(<AuthErrorBanner message={null} onDismiss={() => {}} />)
    expect(container.firstChild).toBeNull()
  })
})

