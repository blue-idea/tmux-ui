/**
 * tokenManager 负责免密 URL Token 提取、sessionStorage 会话记忆与地址栏隐私净化
 */

export const AUTH_STORAGE_KEY = 'tmux_ui_auth_token'

let memoryToken: string | null = null

/**
 * 从 URL Query 查询字符串中解析 token 或 auth_token 参数
 */
export function extractTokenFromUrl(searchStr?: string): string | null {
  if (typeof window === 'undefined') return null

  const search = searchStr !== undefined ? searchStr : window.location.search
  if (!search) return null

  const params = new URLSearchParams(search)
  const token = params.get('token') || params.get('auth_token')
  return token ? token.trim() : null
}

/**
 * 净化浏览器地址栏中的敏感 token 参数，防止截图与分享时凭证外泄
 */
export function scrubUrlToken(): void {
  if (typeof window === 'undefined' || !window.history || !window.location) return

  const search = window.location.search
  if (!search) return

  const params = new URLSearchParams(search)
  const hadToken = params.has('token') || params.has('auth_token')
  if (!hadToken) return

  params.delete('token')
  params.delete('auth_token')

  const newQuery = params.toString()
  const newRelativePathQuery =
    window.location.pathname + (newQuery ? `?${newQuery}` : '') + window.location.hash

  try {
    window.history.replaceState({}, document.title, newRelativePathQuery)
  } catch {
    // 忽略特定沙箱环境下的安全限制
  }
}

export interface InitAuthTokenOptions {
  scrubUrl?: boolean
}

/**
 * 初始化并获取当前访问凭证：
 * 1. 优先从 URL Query 提取
 * 2. 提取成功后存入 sessionStorage 并默认净化地址栏
 * 3. 若 URL 无参数，则回退读取 sessionStorage
 */
export function initAuthToken(options: InitAuthTokenOptions = {}): string | null {
  const { scrubUrl = true } = options

  // 1. 优先从当前 URL 提取
  const urlToken = extractTokenFromUrl()
  if (urlToken) {
    setAuthToken(urlToken)
    if (scrubUrl) {
      scrubUrlToken()
    }
    return urlToken
  }

  // 2. 回退从 sessionStorage 获取
  return getAuthToken()
}

/**
 * 获取当前已生效的 Token（优先内存缓存，后查 sessionStorage）
 */
export function getAuthToken(): string | null {
  if (memoryToken) {
    return memoryToken
  }

  if (typeof window !== 'undefined' && window.sessionStorage) {
    try {
      const stored = window.sessionStorage.getItem(AUTH_STORAGE_KEY)
      if (stored) {
        memoryToken = stored.trim()
        return memoryToken
      }
    } catch {
      // 容错处理
    }
  }

  return null
}

/**
 * 更新设置 Token
 */
export function setAuthToken(token: string | null): void {
  memoryToken = token ? token.trim() : null

  if (typeof window !== 'undefined' && window.sessionStorage) {
    try {
      if (memoryToken) {
        window.sessionStorage.setItem(AUTH_STORAGE_KEY, memoryToken)
      } else {
        window.sessionStorage.removeItem(AUTH_STORAGE_KEY)
      }
    } catch {
      // 容错处理
    }
  }
}

/**
 * 清除凭证信息
 */
export function clearAuthToken(): void {
  setAuthToken(null)
}

