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

本工程严格遵循规范驱动开发（SDD）全流程，相关规格资产：

- [产品需求规范 (Requirements)](docs/spec/versions/v1/requirements.md)
- [系统架构与流程设计 (Design)](docs/spec/versions/v1/design.md)
- [数据模型与状态机 (Data)](docs/spec/versions/v1/data.md)
- [REST / WebSocket 接口契约 (API)](docs/spec/versions/v1/api.md)
- [原子任务拆分与执行清单 (Tasks)](docs/spec/versions/v1/tasks.md)
- [端到端可追溯性矩阵 (Traceability)](docs/spec/versions/v1/traceability.md)
- [最终验收与测试交付报告](docs/spec/versions/v1/reports/acceptance-report-2026-11-11.md)
