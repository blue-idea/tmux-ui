# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v1/ac/TASK-012-AC.md`
> 适用版本：v1
> 任务编号：TASK-012
> 执行日期：2026-11-11
> 执行人：Antigravity Agent

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-012 | v1.REQ-001-AC-005 | Build / CLI | 成功编译产出无外部 Node 依赖的单二进制可执行文件 `bin/tmux-ui.exe`（体积约 12MB）；CLI 支持 `--port`、`--token`、`--socket-name`、`--socket-path`、`--config`、`--interval` 等参数并提供 `--help` 使用说明 | PASS | `.\bin\tmux-ui.exe --help` | — |
| TASK-012 | v1.REQ-005-AC-001 | Integration | 运行单二进制实例（端口 8089），成功通过真实 HTTP 请求验证健康检查接口返回 HTTP 200，Token 鉴权保护生效（未携带返回 HTTP 401，携带返回 HTTP 200） | PASS | `curl.exe -i http://127.0.0.1:8089/api/v1/health` & `curl.exe -i http://127.0.0.1:8089/api/v1/panes` | — |
| TASK-012 | v1.REQ-006-AC-001 | Integration / E2E | Go `embed.FS` 完美内嵌前端 React + Vite + Tailwind CSS + xterm.js 静态产物，支持客户端 SPA 路由 Fallback；Playwright 真实导航到单二进制监听端口截取完整运行界面 | PASS | `curl.exe -i http://127.0.0.1:8089/` & Playwright 视觉快照截图 | — |

---

## 真实测试执行输出

### 1. 二进制构建与 CLI 帮助输出

```
Usage of E:\Nextcloud\coding\agents\tmux-ui\bin\tmux-ui.exe:
  -config string
    	Path to agent rules configuration YAML (default "config/agents.yaml")
  -interval int
    	Polling interval in milliseconds (default 1000)
  -port int
    	HTTP server listening port (default 8080)
  -socket-name string
    	Tmux socket name (-L)
  -socket-path string
    	Tmux socket absolute path (-S)
  -token string
    	Optional authentication token for REST API and WebSocket
```

### 2. 真实 HTTP API 与 Token 拦截响应

```
# 健康检查
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 121

{"code":0,"data":{"active_panes":0,"status":"healthy","tmux_version":"unknown","uptime_seconds":22},"message":"success"}

# 鉴权保护拦截
HTTP/1.1 401 Unauthorized
Content-Type: application/json
Content-Length: 89

{"error":{"code":"UNAUTHORIZED","message":"Authentication token is missing or invalid"}}

# 合法 Token 放行
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 41

{"code":0,"data":[],"message":"success"}
```

### 3. 静态页面嵌入验证

```
HTTP/1.1 200 OK
Accept-Ranges: bytes
Content-Length: 528
Content-Type: text/html; charset=utf-8

<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/vite.svg" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>tmux-ui · Multi-Agent Visualizer</title>
    <script type="module" crossorigin src="/assets/index-BWQHPfcW.js"></script>
    <link rel="stylesheet" crossorigin href="/assets/index-CKPKgPOS.css">
  </head>
  <body class="bg-slate-950 text-slate-100 antialiased">
    <div id="root"></div>
  </body>
</html>
```

## 视觉回归测试（Visual Regression Screenshot）

- **截屏工具**：Playwright MCP (`playwright_navigate` + `playwright_screenshot`)
- **证据文件**：`docs/spec/versions/v1/ac/screenshots/TASK-012-standalone-binary.png`
- **页面视图**：单二进制监听端口承载的完整自主多智能体可视化控制台界面。

