import { renderHook, act } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useMediaQuery, useIsMobile } from './useMediaQuery'

describe('useMediaQuery & useIsMobile', () => {
  let listeners: Array<(e: { matches: boolean }) => void> = []
  let matches = false

  beforeEach(() => {
    listeners = []
    matches = false

    window.matchMedia = vi.fn().mockImplementation((query: string) => ({
      matches,
      media: query,
      onchange: null,
      addListener: vi.fn((fn) => listeners.push(fn)),
      removeListener: vi.fn(),
      addEventListener: vi.fn((_, fn) => listeners.push(fn)),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    }))
  })

  it('初始视口匹配返回 false', () => {
    const { result } = renderHook(() => useMediaQuery('(max-width: 767px)'))
    expect(result.current).toBe(false)
  })

  it('useIsMobile 随视口变动响应式更新', () => {
    const { result } = renderHook(() => useIsMobile())
    expect(result.current).toBe(false)

    // 模拟视口变更为移动端
    act(() => {
      matches = true
      listeners.forEach((fn) => fn({ matches: true }))
    })

    expect(result.current).toBe(true)
  })
})
