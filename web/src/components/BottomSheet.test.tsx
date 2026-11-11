import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { BottomSheet } from './BottomSheet'

describe('BottomSheet Component', () => {
  it('当 isOpen 为 false 时不渲染任何内容', () => {
    const { container } = render(
      <BottomSheet isOpen={false} onClose={vi.fn()} title="Test Title">
        <p>Content</p>
      </BottomSheet>
    )
    expect(container.firstChild).toBeNull()
  })

  it('当 isOpen 为 true 时展示标题和内容，支持点击关闭按钮与遮罩关闭', () => {
    const onClose = vi.fn()
    render(
      <BottomSheet isOpen={true} onClose={onClose} title="Mobile Timeline">
        <p>Timeline Event 1</p>
      </BottomSheet>
    )

    expect(screen.getByText('Mobile Timeline')).toBeInTheDocument()
    expect(screen.getByText('Timeline Event 1')).toBeInTheDocument()

    // 点击关闭按钮
    const closeBtn = screen.getByRole('button', { name: /close/i })
    fireEvent.click(closeBtn)
    expect(onClose).toHaveBeenCalledTimes(1)

    // 点击背景遮罩
    const backdrop = screen.getByTestId('bottom-sheet-backdrop')
    fireEvent.click(backdrop)
    expect(onClose).toHaveBeenCalledTimes(2)
  })
})
