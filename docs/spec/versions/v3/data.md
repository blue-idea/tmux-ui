# 数据与配置模型设计（Data）

> 文件路径：`docs/spec/versions/v3/data.md`
> 适用版本：v3 · 生成日期：2026-11-12
> 前置宪章：沿用 `docs/spec/versions/v1-finished/constitution.md` 与项目级 Steering 规则

---

## 1. Cloudflare Tunnel 与公网配置模型

### 1.1 服务端隧道配置结构（Go 后端）
在 `internal/model/config.go` 中扩展 `TunnelConfig` 模型：

```go
// TunnelConfig 定义 Cloudflare Tunnel 与公网暴露选项
type TunnelConfig struct {
	Enabled       bool   `yaml:"enabled" json:"enabled"`                 // 是否启用 Tunnel 自动管理
	PublicURL     string `yaml:"public_url" json:"public_url"`           // 自定义域名或已有公网地址
	Cloudflared   string `yaml:"cloudflared_bin" json:"cloudflared_bin"` // cloudflared 可执行文件路径（默认系统 PATH）
	AutoGenerateToken bool `yaml:"auto_generate_token" json:"auto_generate_token"` // 无 token 时是否自动生成安全 token
}
```

### 1.2 命令行与环境变量映射

| 配置项 | CLI 参数 | 环境变量 | 默认值 | 说明 |
|--------|---------|---------|--------|------|
| 隧道启用 | `--tunnel` | `TMUX_UI_TUNNEL` | `false` | 启用后若未设置 token 则自动生成随机安全密钥 |
| 公网地址 | `--public-url` | `TMUX_UI_PUBLIC_URL` | `""` | 指定已配置好的隧道公网 URL |
| 认证令牌 | `--token` | `TMUX_UI_TOKEN` | `""` | 访问密钥，开启公网时强校验 |
| cloudflared 路径 | `--cloudflared-bin` | `CLOUDFLARED_BIN` | `"cloudflared"` | 可执行文件路径 |

---

## 2. 前端客户端认证凭证存储模型（TypeScript 前端）

### 2.1 存储机制设计
- **主存储媒介**：`sessionStorage`（键名：`tmux_ui_auth_token`）
  - 理由：在 Telegram Mini App、飞书、微信等内嵌 WebView 以及普通移动浏览器标签页生命周期内稳定存在，关闭标签页自动回收，避免 localStorage 永久残留在共享设备或公共电脑上的潜在安全风险。
- **降级回退机制**：当 URL Query 包含 `token` 或 `auth_token` 时，优先读取并立即同步至 `sessionStorage`；若 URL 中无参数，自动回退到 `sessionStorage` 中已持久化的 Token。

### 2.2 前端认证状态结构
在 `web/src/types/auth.ts` 中定义：

```typescript
export interface AuthState {
  token: string | null
  isAuthenticated: boolean
  authError: string | null
}
```

