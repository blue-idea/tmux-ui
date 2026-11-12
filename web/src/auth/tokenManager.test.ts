import { describe, it, expect, beforeEach, vi } from 'vitest'
import {
  initAuthToken,
  getAuthToken,
  setAuthToken,
  clearAuthToken,
  extractTokenFromUrl,
  scrubUrlToken,
} from './tokenManager'

describe('tokenManager 凭据与免密秒登管理', () => {
  beforeEach(() => {
    // 重置 sessionStorage 与历史记录
    sessionStorage.clear()
    clearAuthToken()
    window.history.replaceState({}, '', '/')
  })

  it('v3.REQ-002-AC-001: 自动从 URL Query 中的 ?token= 提取凭证并暂存至 sessionStorage', () => {
    window.history.replaceState({}, '', '/?token=test_secret_token_123')

    const token = initAuthToken({ scrubUrl: false })
    expect(token).toBe('test_secret_token_123')
    expect(getAuthToken()).toBe('test_secret_token_123')
    expect(sessionStorage.getItem('tmux_ui_auth_token')).toBe('test_secret_token_123')
  })

  it('v3.REQ-002-AC-001: 支持兼容命名 ?auth_token= 参数提取', () => {
    window.history.replaceState({}, '', '/?auth_token=compat_token_456')

    const token = initAuthToken({ scrubUrl: false })
    expect(token).toBe('compat_token_456')
    expect(getAuthToken()).toBe('compat_token_456')
  })

  it('v3.REQ-002-AC-002: 成功提取 Token 后自动调用 replaceState 净化地址栏且保留其他参数', () => {
    const replaceStateSpy = vi.spyOn(window.history, 'replaceState')
    window.history.replaceState({}, '', '/?token=secret999&filter=active')

    const token = initAuthToken({ scrubUrl: true })
    expect(token).toBe('secret999')

    expect(replaceStateSpy).toHaveBeenCalled()
    expect(window.location.search).toBe('?filter=active')
    expect(window.location.search).not.toContain('secret999')
    replaceStateSpy.mockRestore()
  })

  it('v3.REQ-002-AC-003: 当 URL 无 Token 参数时，回退读取 sessionStorage 中的已持久化 Token', () => {
    sessionStorage.setItem('tmux_ui_auth_token', 'cached_session_token_888')
    window.history.replaceState({}, '', '/?view=grid')

    const token = initAuthToken()
    expect(token).toBe('cached_session_token_888')
    expect(getAuthToken()).toBe('cached_session_token_888')
  })

  it('clearAuthToken 应同时清空内存缓存与 sessionStorage', () => {
    setAuthToken('temp_token')
    expect(getAuthToken()).toBe('temp_token')
    expect(sessionStorage.getItem('tmux_ui_auth_token')).toBe('temp_token')

    clearAuthToken()
    expect(getAuthToken()).toBeNull()
    expect(sessionStorage.getItem('tmux_ui_auth_token')).toBeNull()
  })
})
