# tmux-ui · Autonomous Agent Monitoring Console

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)](https://golang.org)
[![React](https://img.shields.io/badge/React-18.3-61DAFB?style=flat&logo=react)](https://react.dev)
[![Vite](https://img.shields.io/badge/Vite-6.0-646CFF?style=flat&logo=vite)](https://vitejs.dev)
[![Tailwind CSS](https://img.shields.io/badge/Tailwind-3.4-38B2AC?style=flat&logo=tailwind-css)](https://tailwindcss.com)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`tmux-ui` 是一款专为管理多个后台自主智能体（Claude Code、Codex、Aider 等）打造的轻量级、开箱即用的可视化控制台。服务端基于 Go 语言构建，前端集成 React 与 xterm.js，打包后为无外部依赖的单一可执行文件。

---

## ✨ 核心特性

- 🖥️ **只读终端监控**：集成 xterm.js，完整保留 ANSI 彩色字符流与终端格式，严格防篡改（只读无 `send-keys`）。
- 🤖 **多智能体状态研判**：
  - 启发式决策链：`pane_dead` 优先 -> `waiting_input` 等待输入正则 -> `error` 报错正则 -> `working` 运行中正则 -> 输出哈希活跃度 -> 纯 Shell `idle`。
  - 内置开箱即用支持：Claude Code、Codex CLI、Aider、Gemini CLI 及纯 Shell。
- ⚡ **毫秒级实时推流**：基于 WebSocket Hub 架构，新连接推送全量快照，后续仅增量广播状态变迁事件与终端输出差异。
- ⏱️ **时间线追溯与系统通知**：支持保留单 Pane 最近 100 条状态变迁历史；智能体进入等待输入或报错时自动触发浏览器系统级桌面通知。
- 📦 **零依赖单二进制交付**：Go `embed.FS` 内嵌编译产物，下载即用，无需配置 Node.js、pnpm 或外部数据库。

---

## 🚀 快速上手

### 1. 运行预编译单二进制

```bash
# 默认监听 127.0.0.1:8080
./bin/tmux-ui

# 指定自定义端口与安全 Token
./bin/tmux-ui --port 9000 --token my-secret-token

# 关联自定义 tmux socket
./bin/tmux-ui --socket-name dev_session
```

打开浏览器访问 `http://127.0.0.1:8080` 即可查看控制台。

---

## 💻 命令行实时状态查询 (CLI Mode)

无需打开浏览器，直接在终端中通过命令行一键获取所有 tmux 会话及智能体的实时执行状态。支持本地直接探测与远程已运行服务端拉取双模式。

### 1. 常用命令速查

```bash
# 1. 终端 ASCII 表格查看实时状态（本地直接采集或直连当前 tmux 会话）
./bin/tmux-ui status

# 2. 支持别名快捷命令 (ps / ls)
./bin/tmux-ui ps
./bin/tmux-ui ls

# 3. 输出结构化 JSON 格式（便于脚本、CI/CD 或告警机器人消费）
./bin/tmux-ui status --json

# 4. 远程/本地向已运行的 tmux-ui 服务端拉取状态（支持 Bearer Token 鉴权）
./bin/tmux-ui status --server http://127.0.0.1:8080 --token my-secret-token

# 5. 指定自定义 tmux socket 名称或绝对路径
./bin/tmux-ui status --socket-name dev_session
./bin/tmux-ui status --socket-path /tmp/tmux-1000/default
```

### 2. 终端表格输出效果

执行 `tmux-ui status` 时的格式化 ASCII 输出效果：

```text
PANE ID   SESSION        AGENT         STATE           COMMAND   IDLE   LAST PREVIEW
-------   -------        -----         -----           -------   ----   ------------
%0        dev:editor     claude-code   [WORKING]       node      2s     Thinking...
%1        prod:agent     aider         [WAITING_INPUT] python    18s    Apply these changes? (y/n)
%2        test:shell     -             [IDLE]          zsh       45s    git status
%3        build:worker   codex         [ERROR]         codex     5s     Process exited with code 1
```

### 3. JSON 格式输出效果

执行 `tmux-ui status --json` 时输出结构化数据：

```json
[
  {
    "pane_id": "%0",
    "session_name": "dev",
    "window_name": "editor",
    "agent_type": "claude-code",
    "state": "working",
    "current_command": "node",
    "idle_seconds": 2,
    "preview_lines": [
      "Running semantic analysis...",
      "Thinking..."
    ]
  }
]
```

### 4. 命令行参数选项速查

| 参数标志 | 简写 / 环境变量 | 默认值 | 说明 |
|---------|---------------|--------|------|
| `--json` | — | `false` | 以结构化 JSON 数组格式输出状态 |
| `--server` | `TMUX_UI_SERVER` | `""` | 目标 tmux-ui 服务端地址（如 `http://127.0.0.1:8080`），指定时走 HTTP API 拉取 |
| `--token` | `TMUX_UI_TOKEN` | `""` | 服务端 API 认证密钥（携带 `Authorization: Bearer <token>`） |
| `--socket-name` | — | `""` | tmux 本地 socket 别名 (`-L`) |
| `--socket-path` | — | `""` | tmux 本地 socket 绝对路径 (`-S`) |
| `--config` | — | `config/agents.yaml` | 智能体识别规则配置文件路径 |

---

## 🌐 Cloudflare Tunnel 公网发布与 ChatOps 免密直登

无需公网 IP 和端口映射，支持配合 Cloudflare Tunnel 将监控台发布至公网，并在 Telegram Mini App、飞书、Discord 等应用中实现点击直登。

### 1. 启动公网隧道模式与自动安全 Token
```bash
# 启动公网隧道模式（未指定 token 时系统自动生成 32 字节高熵防裸奔随机密钥）
./bin/tmux-ui --tunnel --public-url https://tmux.yourdomain.com
```

终端将打印高亮免密直登卡片：
```text
┌────────────────────────────────────────────────────────────────────────────┐
│                        tmux-ui Cloudflare Tunnel Ready                     │
├────────────────────────────────────────────────────────────────────────────┤
│ One-Click Public Login: https://tmux.yourdomain.com/?token=3d2984b6...     │
│ Secret Auth Token:     3d2984b6...                                         │
└────────────────────────────────────────────────────────────────────────────┘
```

### 2. 内嵌应用无感秒登与防泄露机制
- **Telegram Mini App / 飞书直登**：点击链接 `https://tmux.yourdomain.com/?token=...` 即可自动进入控制台。
- **全链路自动鉴权**：前端自动从 URL 提取 Token 注入 WebSocket 与 REST API，并暂存至 `sessionStorage` 保证刷新不丢态。
- **地址栏隐私净化**：成功提取后自动通过 `history.replaceState` 清除 URL 中的敏感 token 参数，防止用户在群内分享或截图时凭证意外泄露。
- 详细生产与快速临时隧道配置参考：[`docs/cloudflare-tunnel.md`](docs/cloudflare-tunnel.md) 与 [`config/tunnel.yml.example`](config/tunnel.yml.example)。

---

### 2. 源码构建

#### 编译前端与后端（产出单二进制）

```bash
# 1. 构建前端静态资源
cd web
pnpm install
pnpm build
cd ..

# 2. 构建独立单二进制程序
go build -o bin/tmux-ui ./cmd/tmux-ui
```

### 3. 运行测试套件

```bash
# 运行后端单元测试与覆盖率统计（要求 > 85%）
go test -v -cover ./...

# 运行前端 Vitest 单元与组件测试
cd web && pnpm test
```

---

## ⚙️ 配置文件说明 (`config/agents.yaml`)

```yaml
collector:
  socket_name: ""            # 可选 tmux -L socket 名称
  socket_path: ""            # 可选 tmux -S socket 绝对路径
  poll_interval_ms: 800      # 采集轮询周期（毫秒）
  capture_tail_lines: 50     # 捕获末尾终端输出行数
  max_buffer_bytes: 524288   # 输出缓冲区大小限制（512KB）
  inactivity_timeout_s: 15   # 超时判定空闲阈值（秒）

agents:
  - name: "claude-code"
    match_process:
      - "claude"
      - "node"
    states:
      working:
        tail_regex: "esc to interrupt|Thinking\\.\\.\\."
      waiting_input:
        tail_regex: "Do you want to proceed\\?|\\(y/n\\)|Allow this action\\?"
      error:
        tail_regex: "Error:|fatal:|Process exited with code [1-9]"
```

---

## 📚 规格资产库 (Spec-Driven Development)

本工程严格遵循规范驱动开发（SDD）全流程，多版本规格资产归档：

- [版本总索引与当前活跃版本 (spec/README.md)](docs/spec/README.md)
- **当前活跃版本 v3**：
  - [v3 产品需求规范 (Requirements)](docs/spec/versions/v3/requirements.md)
  - [v3 系统架构与流程设计 (Design)](docs/spec/versions/v3/design.md)
  - [v3 数据模型与状态机 (Data)](docs/spec/versions/v3/data.md)
  - [v3 REST / WebSocket 接口契约 (API)](docs/spec/versions/v3/api.md)
  - [v3 原子任务拆分与执行清单 (Tasks)](docs/spec/versions/v3/tasks.md)
  - [v3 端到端可追溯性矩阵 (Traceability)](docs/spec/versions/v3/traceability.md)
  - [v3 最终验收与综合测试交付报告](docs/spec/versions/v3/reports/acceptance-report-2026-11-12.md)
- **已完成历史版本**：
  - [v1 基础拓扑采集与 Web 监控 (v1-finished)](docs/spec/versions/v1-finished/)
  - [v2 历史快照持久化与移动端响应式 (v2-finished)](docs/spec/versions/v2-finished/)
