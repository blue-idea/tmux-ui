# 可追溯性矩阵（Traceability）

> 文件路径：`docs/spec/versions/v1/traceability.md`
> 适用版本：v1 · 生成日期：2026-11-11
> 创建步骤：STEP 5（拆分任务，初始化骨架）
> 更新步骤：STEP 6（每项 TASK 完成后更新状态）

---

## 矩阵说明

| 字段 | 含义 |
|------|------|
| `TASK ID` | 任务编号，对应 `tasks.md` |
| `REQ / AC` | 关联需求编号和验收标准编号（带版本前缀） |
| `状态` | `待开始` / `进行中` / `done` / `BLOCKED` / `WONTFIX` |
| `实现文件` | 主要实现代码的文件路径（完成后填写） |
| `测试文件` | 对应测试文件路径（完成后填写） |
| `AC 报告` | 验收矩阵文件路径，如 `docs/spec/versions/v1/ac/TASK-001-AC.md` |
| `完成日期` | 任务关闭日期 |

---

## 追溯矩阵

| TASK ID | 任务名称 | REQ / AC | 状态 | 实现文件 | 测试文件 | AC 报告 | 完成日期 |
|---------|---------|---------|:----:|---------|---------|---------|---------|
| TASK-001 | 项目基础脚手架、配置模型与解析器 | v1.REQ-004 / AC-001, AC-002 | done | `internal/config/config.go`, `internal/model/*.go` | `internal/config/config_test.go`, `internal/model/state_test.go` | `docs/spec/versions/v1/ac/TASK-001-AC.md` | 2026-11-11 |
| TASK-002 | Tmux CLI 采集器与版本探测兼容 | v1.REQ-001 / AC-001, AC-003, AC-004, AC-005 | done | `internal/collector/tmux_cli.go` | `internal/collector/tmux_cli_test.go` | `docs/spec/versions/v1/ac/TASK-002-AC.md` | 2026-11-11 |
| TASK-003 | 跨平台进程树探测器与 Agent 进程关联 | v1.REQ-001 / AC-002 | done | `internal/collector/process_tree.go` | `internal/collector/process_tree_test.go` | `docs/spec/versions/v1/ac/TASK-003-AC.md` | 2026-11-11 |
| TASK-004 | 终端 ANSI 输出捕获与哈希差异跟踪 | v1.REQ-002 / AC-001, AC-002, AC-003 | done | `internal/collector/capture_pane.go` | `internal/collector/capture_pane_test.go` | `docs/spec/versions/v1/ac/TASK-004-AC.md` | 2026-11-11 |
| TASK-005 | 多智能体状态判定引擎与内存存储 | v1.REQ-003 / AC-001, AC-002, AC-003, AC-004, AC-005 | done | `internal/detector/engine.go`, `internal/store/*.go` | `internal/detector/engine_test.go`, `internal/store/memory_store_test.go` | `docs/spec/versions/v1/ac/TASK-005-AC.md` | 2026-11-11 |
| TASK-006 | 后端 HTTP REST API 与 Token 认证中间件 | v1.REQ-005 / AC-001, AC-002, AC-004 | done | `internal/server/http_server.go` | `internal/server/http_server_test.go` | `docs/spec/versions/v1/ac/TASK-006-AC.md` | 2026-11-11 |
| TASK-007 | WebSocket Hub 实时推流服务 | v1.REQ-005 / AC-003 | done | `internal/server/ws_hub.go`, `internal/server/http_server.go` | `internal/server/ws_hub_test.go` | `docs/spec/versions/v1/ac/TASK-007-AC.md` | 2026-11-11 |
| TASK-008 | 前端 Vite 脚手架与 WebSocket 通信 Hook | v1.REQ-006 / AC-001 | done | `web/src/hooks/useTmuxWS.ts`, `web/src/types/index.ts` | `web/src/hooks/useTmuxWS.test.ts` | `docs/spec/versions/v1/ac/TASK-008-AC.md` | 2026-11-11 |
| TASK-009 | Web 总览卡片墙与多维状态筛选 | v1.REQ-006 / AC-001, AC-002 | done | `web/src/components/OverviewGrid.tsx`, `web/src/components/PaneCard.tsx`, `web/src/components/FilterBar.tsx` | `web/src/components/OverviewGrid.test.tsx` | `docs/spec/versions/v1/ac/TASK-009-AC.md` | 2026-11-11 |
| TASK-010 | 终端详情页与 xterm.js ANSI 渲染器 | v1.REQ-007 / AC-001, AC-003 | done | `web/src/components/TerminalViewer.tsx` | `web/src/components/TerminalViewer.test.tsx` | `docs/spec/versions/v1/ac/TASK-010-AC.md` | 2026-11-11 |
| TASK-011 | 状态变迁时间线抽屉与浏览器交互通知 | v1.REQ-006, v1.REQ-007 / AC-003, AC-002 | done | `web/src/components/TimelineDrawer.tsx`, `web/src/hooks/useNotifications.ts` | `web/src/components/TimelineDrawer.test.tsx`, `web/src/hooks/useNotifications.test.ts` | `docs/spec/versions/v1/ac/TASK-011-AC.md` | 2026-11-11 |
| TASK-012 | 前端静态资源 Embed 内嵌与单二进制集成验收 | v1.REQ-001, v1.REQ-005, v1.REQ-006 / AC-005, AC-001, AC-001 | done | `cmd/tmux-ui/main.go`, `web/embed.go`, `internal/server/http_server.go` | `internal/server/http_server_test.go` | `docs/spec/versions/v1/ac/TASK-012-AC.md` | 2026-11-11 |

> 每完成一项 TASK，将状态改为 `done`，并填入实现文件、测试文件和 AC 报告路径。

---

## 变更记录

| 日期 | 变更内容 | 原因 |
|------|---------|------|
| 2026-11-11 | 初始化矩阵（共 12 项任务） | STEP 5 任务拆分完成 |

