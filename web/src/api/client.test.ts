import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { fetchWithAuth, apiClient } from './client'
import { setAuthToken, clearAuthToken } from '../auth/tokenManager'

describe('HTTP API 统一客户端与自动鉴权 (client.ts)', () => {
  const originalFetch = global.fetch

  beforeEach(() => {
    clearAuthToken()
    global.fetch = vi.fn()
  })

  afterEach(() => {
    global.fetch = originalFetch
  })

  it('v3.REQ-003-AC-002: 当存在已保存 Token 时，请求自动附加 Authorization: Bearer <token> 请求头', async () => {
    setAuthToken('my_test_auth_token_777')

    const mockResponse = new Response(JSON.stringify({ code: 0, data: { status: 'healthy' } }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    })
    ;(global.fetch as any).mockResolvedValue(mockResponse)

    await fetchWithAuth('/api/v1/health')

    expect(global.fetch).toHaveBeenCalledTimes(1)
    const [callUrl, callInit] = (global.fetch as any).mock.calls[0]
    expect(callUrl).toBe('/api/v1/health')
    expect(callInit.headers.get('Authorization')).toBe('Bearer my_test_auth_token_777')
  })

  it('当未配置 Token 时，正常发起请求且不附加无效 Authorization 头', async () => {
    const mockResponse = new Response(JSON.stringify({ code: 0 }), { status: 200 })
    ;(global.fetch as any).mockResolvedValue(mockResponse)

    await fetchWithAuth('/api/v1/health')

    const [, callInit] = (global.fetch as any).mock.calls[0]
    expect(callInit.headers.get('Authorization')).toBeNull()
  })

  it('apiClient.getPaneDetail 能正确编码 Pane ID 并附加鉴权', async () => {
    setAuthToken('auth_token_abc')
    const mockResponse = new Response(
      JSON.stringify({ code: 0, data: { pane_id: '%1', state: 'working' } }),
      { status: 200 }
    )
    ;(global.fetch as any).mockResolvedValue(mockResponse)

    const data = await apiClient.getPaneDetail('%1')
    expect(data.pane_id).toBe('%1')

    const [callUrl, callInit] = (global.fetch as any).mock.calls[0]
    expect(callUrl).toBe('/api/v1/panes/%251')
    expect(callInit.headers.get('Authorization')).toBe('Bearer auth_token_abc')
  })

  it('apiClient.getSnapshots 能够正确请求快照列表', async () => {
    setAuthToken('auth_token_abc')
    const mockResponse = new Response(
      JSON.stringify({ code: 0, data: [{ snapshot_id: 'snap1', timestamp: 12345 }] }),
      { status: 200 }
    )
    ;(global.fetch as any).mockResolvedValue(mockResponse)

    const snapshots = await apiClient.getSnapshots('%1')
    expect(snapshots).toHaveLength(1)
    expect(snapshots[0].snapshot_id).toBe('snap1')

    const [callUrl] = (global.fetch as any).mock.calls[0]
    expect(callUrl).toBe('/api/v1/panes/%251/snapshots')
  })

  it('当响应状态为 401 UNAUTHORIZED 时抛出标准鉴权错误', async () => {
    setAuthToken('expired_token')
    const mockResponse = new Response(
      JSON.stringify({
        error: { code: 'UNAUTHORIZED', message: 'Authentication token is missing or invalid' },
      }),
      { status: 401 }
    )
    ;(global.fetch as any).mockResolvedValue(mockResponse)

    await expect(fetchWithAuth('/api/v1/panes/%251')).rejects.toThrow('Authentication token is missing or invalid')
  })
})
