import { getAuthToken } from '../auth/tokenManager'
import type { PaneDetail, SnapshotMetadata, SnapshotRecord } from '../types'

export type AuthErrorListener = (message: string) => void

const authErrorListeners = new Set<AuthErrorListener>()

/**
 * 注册全局 401 鉴权失败监听器
 */
export function addAuthErrorListener(listener: AuthErrorListener): () => void {
  authErrorListeners.add(listener)
  return () => {
    authErrorListeners.delete(listener)
  }
}

/**
 * 触发全局鉴权失败通知
 */
function notifyAuthError(message: string): void {
  authErrorListeners.forEach((listener) => {
    try {
      listener(message)
    } catch {
      // 容错
    }
  })
}

/**
 * fetchWithAuth 统一注入 Bearer 鉴权凭证的 HTTP 请求封装
 */
export async function fetchWithAuth(
  input: RequestInfo | URL,
  init?: RequestInit
): Promise<Response> {
  const headers = new Headers(init?.headers)

  // 自动从 tokenManager 获取当前会话生效的 Token
  const token = getAuthToken()
  if (token && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch(input, {
    ...init,
    headers,
  })

  // 拦截 401 未授权响应
  if (response.status === 401) {
    let errMsg = 'Authentication token is missing or invalid'
    try {
      const cloned = response.clone()
      const json = await cloned.json()
      if (json && json.error && json.error.message) {
        errMsg = json.error.message
      }
    } catch {
      // 容错处理非 JSON 错误体
    }
    notifyAuthError(errMsg)
    throw new Error(errMsg)
  }

  return response
}

/**
 * 结构化业务 API 客户端
 */
export const apiClient = {
  /**
   * 获取目标 Pane 的全量详情数据
   */
  async getPaneDetail(paneId: string): Promise<PaneDetail> {
    const encodedId = encodeURIComponent(paneId)
    const res = await fetchWithAuth(`/api/v1/panes/${encodedId}`)
    if (!res.ok) {
      throw new Error(`Failed to fetch pane detail: ${res.statusText}`)
    }
    const json = await res.json()
    return json.data
  },

  /**
   * 获取指定 Pane 的历史快照元数据列表
   */
  async getSnapshots(paneId: string, since?: number, limit?: number): Promise<SnapshotMetadata[]> {
    const encodedId = encodeURIComponent(paneId)
    const params = new URLSearchParams()
    if (since !== undefined) params.set('since', since.toString())
    if (limit !== undefined) params.set('limit', limit.toString())

    const query = params.toString() ? `?${params.toString()}` : ''
    const res = await fetchWithAuth(`/api/v1/panes/${encodedId}/snapshots${query}`)
    if (!res.ok) {
      throw new Error(`Failed to list snapshots: ${res.statusText}`)
    }
    const json = await res.json()
    return json.data || []
  },

  /**
   * 获取单条快照记录详情
   */
  async getSnapshotDetail(paneId: string, snapshotId: string): Promise<SnapshotRecord> {
    const encodedId = encodeURIComponent(paneId)
    const res = await fetchWithAuth(`/api/v1/panes/${encodedId}/snapshots/${encodeURIComponent(snapshotId)}`)
    if (!res.ok) {
      throw new Error(`Failed to fetch snapshot detail: ${res.statusText}`)
    }
    const json = await res.json()
    return json.data
  },
}
