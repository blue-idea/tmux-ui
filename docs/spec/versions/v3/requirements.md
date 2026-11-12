# 需求文档（Requirements）

> 文件路径：`docs/spec/versions/v3/requirements.md`
> 适用版本：v3 · 生成日期：2026-11-12
> 状态：草稿 / 待确认（Draft）
> 前置宪章：沿用 `docs/spec/versions/v1-finished/constitution.md` 与项目级 Steering 规则

---

## 1. 介绍

### 1.1 项目目标与背景
tmux-ui v1 和 v2 版本构建了本地 tmux 智能体拓扑监控、实时 ANSI 输出解析、状态变迁检测、历史快照回放以及移动端多视口自适应布局。

进入 **v3 版本**，系统重点解决移动跨端访问与公网穿透使用痛点：
1. **Cloudflare Tunnel 公网发布与反向代理适配**：在无需公网独立 IP 和路由器端口映射的情况下，借助 Cloudflare Tunnel 将内网运行的 tmux-ui 服务快速、安全地穿透至公网；服务端适配反向代理头（`CF-Connecting-IP`、`X-Forwarded-For`、`X-Forwarded-Proto`），并在启动时直接生成并打印带鉴权 Token 的公网直达访问链接。
2. **前端免密 URL 参数支持与无感秒登**：用户在 Telegram Mini App、飞书、Discord、Slack 等即时通讯应用的 Webview/内嵌浏览器中，点击包含 Token 的 URL（如 `https://tunnel.example.com/?token=xxx`）即可无感秒登，自动将 Token 注入到 WebSocket 连接与所有 REST API 请求中；提取后自动净化地址栏以防分享泄露，并在会话期间持久记忆凭据。

### 1.2 核心用户角色
- **跨平台即时通讯监控者 (ChatOps / Mobile Agent Operator)**：在 Telegram、飞书、Discord、Slack 等外部通讯工具中接收智能体告警通知，一键点击通知中的直达链接无感进入控制台查看，无需重新输入密码或繁琐登录。
- **免公网 IP 远程开发者 (Remote Developer / Homelab User)**：在私有服务器或本地开发机部署智能体，通过 Cloudflare Tunnel 快速穿透，随时随地在任何网络环境下安全连回监控终端。

---

## 2. 需求列表

---

### 需求 v3.REQ-001 · Cloudflare Tunnel 公网发布与反向代理安全适配

**用户故事：** 作为远程开发者，我希望通过 Cloudflare Tunnel 将本地 tmux-ui 暴露到公网，并在启动时获得带安全 Token 的完整公网访问链接，以便安全便捷地在任意网络下访问监控台。

#### 验收标准

```yaml
- id: v3.REQ-001-AC-001
  ears: >
    When the server starts with tunnel mode enabled (--tunnel or TMUX_UI_TUNNEL=true),
    the CLI shall initiate or detect cloudflared tunnel integration and print the public access URL with the authentication token query parameter.
  test_type: Unit
  expected:
    return_value: "Public URL formatted with https protocol and ?token=... query string"
    throws: ""
    side_effects:
      - "Log console prints copy-paste ready public access URL"

- id: v3.REQ-001-AC-002
  ears: >
    When a request traverses Cloudflare Tunnel with CF-Connecting-IP or X-Forwarded-For headers,
    the HTTP server shall correctly resolve and record the client's public IP address instead of the local tunnel loopback IP.
  test_type: API
  expected:
    http_status: 200
    side_effects:
      - "RealIP middleware extracts CF-Connecting-IP as remote client IP"

- id: v3.REQ-001-AC-003
  ears: >
    When public tunnel mode is enabled and no authentication token is explicitly configured by the user,
    the server shall automatically generate a cryptographically secure random token (>= 32 characters) and log a prominent security warning, refusing unauthenticated public exposure.
  test_type: Unit
  expected:
    return_value: "Generated secure token length >= 32"
    throws: ""
    side_effects:
      - "Server enforces authentication token on all protected endpoints"

- id: v3.REQ-001-AC-004
  ears: >
    When deploying Cloudflare Tunnel in production,
    the project shall provide standard configuration templates (config/tunnel.yml.example) and deployment documentation for both Quick Tunnels and Named Tunnels with WebSocket ingress rules.
  test_type: Manual
  expected:
    side_effects:
      - "config/tunnel.yml.example and docs/cloudflare-tunnel.md exist and validate with cloudflared schema"
```

---

### 需求 v3.REQ-002 · 前端免密 URL 参数提取与会话凭据管理

**用户故事：** 作为 ChatOps 运维者，我希望在 Telegram Mini App 或飞书等应用内点击带 Token 的链接时，前端自动提取 Token 并不需二次确认直接登录，同时从浏览器地址栏擦除 Token 以防泄露。

#### 验收标准

```yaml
- id: v3.REQ-002-AC-001
  ears: >
    When the web application loads with a "token" or "auth_token" query parameter in the window URL,
    the auth module shall immediately extract the token, persist it into sessionStorage, and make it globally available for authentication.
  test_type: Unit
  expected:
    return_value: "Extracted valid token string"
    side_effects:
      - "sessionStorage.getItem('tmux_ui_token') equals the extracted token"

- id: v3.REQ-002-AC-002
  ears: >
    When a token is successfully extracted from the URL query parameters,
    the application shall call window.history.replaceState to purge the token parameter from the browser address bar without triggering a page reload.
  test_type: Unit
  expected:
    side_effects:
      - "window.location.search no longer contains token or auth_token"
      - "Clean URL preserved in address bar"

- id: v3.REQ-002-AC-003
  ears: >
    When the web application is refreshed or navigated within the same browser session without a token in the URL,
    the auth module shall fallback to reading the token from sessionStorage.
  test_type: Unit
  expected:
    return_value: "Cached token string from sessionStorage"
```

---

### 需求 v3.REQ-003 · WebSocket 与 REST API 全链路凭据自动注入

**用户故事：** 作为前端开发者，我希望 WebSocket 连接与所有 REST API 请求在发起时统一、自动地注入提取到的 Token，使得终端输出、详情查询和历史快照拉取均能无缝完成鉴权。

#### 验收标准

```yaml
- id: v3.REQ-003-AC-001
  ears: >
    When useTmuxWS hook initializes without explicit options.token,
    it shall automatically retrieve the token from the auth module and append "?token=..." to the WebSocket handshake URL.
  test_type: Unit
  expected:
    side_effects:
      - "WebSocket instance connects to target URL containing ?token=<stored_token>"

- id: v3.REQ-003-AC-002
  ears: >
    When any REST API request is dispatched (including pane details and historical snapshots),
    the unified API client shall automatically inject the stored token into the HTTP "Authorization: Bearer <token>" header (or fallback query parameter).
  test_type: Unit
  expected:
    side_effects:
      - "Outgoing fetch request carries Authorization: Bearer <stored_token>"

- id: v3.REQ-003-AC-003
  ears: >
    When the backend rejects a WebSocket connection or REST API request with 401 UNAUTHORIZED,
    the web UI shall display an authentication error notification banner prompting the user that their token is missing or expired.
  test_type: Unit
  expected:
    side_effects:
      - "Auth error banner or toast rendered in UI with clear English message"
```

---

### 需求 v3.REQ-004 · CLI 实时状态查询与终端渲染

**用户故事：** 作为习惯于命令行工作的运维者或开发者，我希望通过命令行子命令直接在终端快速查看当前所有 tmux 会话与智能体的实时运行状态，无需打开浏览器。

#### 验收标准

```yaml
- id: v3.REQ-004-AC-001
  ears: >
    When the user executes "tmux-ui status",
    the CLI shall inspect the active tmux panes and output a well-aligned ASCII table displaying Pane ID, Session, Agent Type, State, Current Command, Idle Time, and Preview.
  test_type: Unit
  expected:
    return_value: "Formatted ASCII table with headers and pane rows"
    throws: ""

- id: v3.REQ-004-AC-002
  ears: >
    When the user executes "tmux-ui status --json",
    the CLI shall output valid JSON array format of all pane statuses for script automation.
  test_type: Unit
  expected:
    return_value: "Valid JSON array matching []PaneSummary schema"
    throws: ""

- id: v3.REQ-004-AC-003
  ears: >
    When the user specifies "--server <URL>",
    the CLI shall query the running tmux-ui server REST API /api/v1/panes (with optional --token Bearer authentication);
    when --server is omitted, it shall directly execute local tmux inspection.
  test_type: Unit
  expected:
    return_value: "Pane status retrieved successfully from remote server or local collector"
    throws: ""
```


