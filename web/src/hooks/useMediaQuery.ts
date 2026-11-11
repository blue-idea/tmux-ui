import { useState, useEffect } from 'react'

/**
 * useMediaQuery 侦测媒体查询匹配状态，支持 SSR 优雅降级与动态监听
 */
export function useMediaQuery(query: string): boolean {
  const [matches, setMatches] = useState<boolean>(() => {
    if (typeof window !== 'undefined' && window.matchMedia) {
      return window.matchMedia(query).matches
    }
    return false
  })

  useEffect(() => {
    if (typeof window === 'undefined' || !window.matchMedia) {
      return
    }

    const mediaQueryList = window.matchMedia(query)
    const listener = (event: MediaQueryListEvent | { matches: boolean }) => {
      setMatches(event.matches)
    }

    setMatches(mediaQueryList.matches)

    if (mediaQueryList.addEventListener) {
      mediaQueryList.addEventListener('change', listener as EventListener)
      return () => mediaQueryList.removeEventListener('change', listener as EventListener)
    } else if (mediaQueryList.addListener) {
      // 兼容旧浏览器
      mediaQueryList.addListener(listener as (e: MediaQueryListEvent) => void)
      return () => mediaQueryList.removeListener(listener as (e: MediaQueryListEvent) => void)
    }
  }, [query])

  return matches
}

/**
 * useIsMobile 便捷 Hook，判断视口宽度是否小于 768px（移动端）
 */
export function useIsMobile(): boolean {
  return useMediaQuery('(max-width: 767px)')
}
