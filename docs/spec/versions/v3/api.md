# 接口与网络协议规范（API）

> 文件路径：`docs/spec/versions/v3/api.md`
> 适用版本：v3 · 生成日期：2026-11-12
> 基础继承：沿用 `docs/spec/versions/v1-finished/api.md` 与 `docs/spec/versions/v2-finished/api.md`

---

## 1. 认证接口协议约定

### 1.1 前端 URL 免密直登入参规范
支持从页面 URL Query 中解析以下两组同义参数：
- `?token=<SECRET_TOKEN>` （主推荐命名）
- `?auth_token=<SECRET_TOKEN>` （兼容命名）

示例直登链接：
```
https://agents.example.com/?token=a8f9c1e204b789d31f0e4c6a8b7c9e12
```

### 1.2 REST API 请求头鉴权规范
所有由前端向后端 `/api/v1/*` 保护端点发起的请求，统一在 HTTP Headers 中注入 Bearer Token：

```http
GET /api/v1/panes/%251 HTTP/1.1
Host: agents.example.com
Authorization: Bearer a8f9c1e204b789d31f0e4c6a8b7c9e12
```

后端保留 URL Query `?token=...` 降级兼容支持，以适配无自定义 Header 能力的特殊环境。

### 1.3 WebSocket 升级握手鉴权
由于浏览器原生的 `WebSocket` API 不支持自定义 HTTP 请求头（如 `Authorization`），WebSocket 连接统一采用 Query 参数传输：

```
wss://agents.example.com/ws?token=a8f9c1e204b789d31f0e4c6a8b7c9e12
```

---

## 2. 反向代理与 Cloudflare Tunnel 协议头

服务端通过 `RealIP` 中间件与反向代理适配，按以下优先级提取客户端真实 IP：
1. `CF-Connecting-IP`（Cloudflare 官方边缘节点注入的访客公网真实 IP）
2. `X-Forwarded-For`（最左侧原始客户端 IP）
3. `r.RemoteAddr`（直连网络 IP）

### 协议自动探测
- 前端通过 `window.location.protocol === 'https:' ? 'wss:' : 'ws:'` 自动根据当前隧道或页面的加密状态适配 WebSocket 握手协议，杜绝在 HTTPS 域名下发起不安全的 `ws://` 导致的 Mixed Content 阻断。

---

## 3. 错误响应码

当 Token 缺失、错误或失效时，服务端与前端统一遵循：

### 401 UNAUTHORIZED
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Authentication token is missing or invalid"
  }
}
```
前端捕获该响应后，激活全局认证异常横幅并提示用户凭证无效。

