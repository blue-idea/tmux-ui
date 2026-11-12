# 可追溯性矩阵（Traceability）

> 文件路径：`docs/spec/versions/v3/traceability.md`
> 适用版本：v3 · 日期：2026-11-12
> 创建步骤：STEP 5（拆分任务，生成初始骨架）

---

## 1. 矩阵说明

| 字段 | 含义 |
|------|------|
| `TASK ID` | 任务编号，对应 `tasks.md` |
| `REQ / AC` | 关联需求编号和验收标准编号（带 v3 前缀） |
| `状态` | `待开始` / `进行中` / `done` / `BLOCKED` / `WONTFIX` |
| `实现文件` | 主要实现代码的文件路径（完成后填写） |
| `测试文件` | 对应测试文件路径（完成后填写） |
| `AC 报告` | 验收矩阵文件路径，如 `docs/spec/versions/v3/ac/TASK-020-AC.md` |
| `完成日期` | 任务关闭日期 |

---

## 2. 追溯矩阵

| TASK ID | 任务名称 | REQ / AC | 状态 | 实现文件 | 测试文件 | AC 报告 | 完成日期 |
|---------|---------|---------|:----:|---------|---------|---------|---------|
| TASK-020 | Cloudflare Tunnel 部署模板、实战文档与反向代理适配 | v3.REQ-001 / AC-002, AC-004 | done | internal/server/http_server.go | internal/server/http_server_test.go | docs/spec/versions/v3/ac/TASK-020-AC.md | 2026-11-12 |
| TASK-021 | 服务端公网隧道集成、安全随机 Token 生成与直登 URL 打印 | v3.REQ-001 / AC-001, AC-003 | done | internal/tunnel/tunnel.go, cmd/tmux-ui/main.go | internal/tunnel/tunnel_test.go | docs/spec/versions/v3/ac/TASK-021-AC.md | 2026-11-12 |
| TASK-022 | 前端 URL Token 解析、sessionStorage 记忆与地址栏隐私净化 | v3.REQ-002 / AC-001, AC-002, AC-003 | 待开始 | — | — | — | — |
| TASK-023 | 前端统一 HTTP API 客户端封装与全链路鉴权请求改造 | v3.REQ-003 / AC-002 | 待开始 | — | — | — | — |
| TASK-024 | 前端 WebSocket 握手 Token 自动链式注入与断线重连 | v3.REQ-003 / AC-001 | 待开始 | — | — | — | — |
| TASK-025 | 鉴权异常横幅组件与全链路集成端到端验收 | v3.REQ-003 / AC-003 | 待开始 | — | — | — | — |

---

## 3. 变更记录

| 日期 | 变更内容 | 原因 |
|------|---------|------|
| 2026-11-12 | 初始化 v3 追溯矩阵（共 6 项任务） | STEP 5 任务拆分完成 |
