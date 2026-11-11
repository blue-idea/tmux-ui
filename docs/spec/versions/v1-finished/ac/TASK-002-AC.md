# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v1/ac/TASK-002-AC.md`
> 适用版本：v1
> 任务编号：TASK-002
> 执行日期：2026-11-11
> 执行人：Antigravity Agent

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-002 | v1.REQ-001-AC-001 | Unit | 成功执行 `list-panes` 并提取 session/window/pane/pid/command/is_dead 等核心字段 | PASS | `go test -v ./internal/collector/...` | — |
| TASK-002 | v1.REQ-001-AC-003 | Unit | 当捕获到 "no server running" 或连接失败信息时，准确转换为类型化错误 `ErrTmuxUnavailable`，无进程崩溃 | PASS | `TestListPanes_ServerUnavailable` | — |
| TASK-002 | v1.REQ-001-AC-004 | Unit | 准确解析 `tmux 3.4`, `tmux 3.2a`, `tmux next-3.5` 等版本格式，异常输出安全回退为 2.0 | PASS | `TestDetectTmuxVersion` | — |
| TASK-002 | v1.REQ-001-AC-005 | Unit | CLI 指定 `--socket-name` / `--socket-path` 时优先覆盖配置文件的 Socket 配置并拼接 `-L`/`-S` | PASS | `TestBuildTmuxArgs_SocketPrecedence` | — |

---

## 真实测试执行输出

```
=== RUN   TestDetectTmuxVersion
=== RUN   TestDetectTmuxVersion/标准版本_3.4
=== RUN   TestDetectTmuxVersion/带字符版本_3.2a
=== RUN   TestDetectTmuxVersion/开发版_next-3.5
=== RUN   TestDetectTmuxVersion/无法识别版本退回默认
=== RUN   TestDetectTmuxVersion/检测版本服务不可用时返回_ErrTmuxUnavailable
=== RUN   TestDetectTmuxVersion/检测版本普通错误
--- PASS: TestDetectTmuxVersion (0.00s)
=== RUN   TestNewTmuxCollector_DefaultExecutor
--- PASS: TestNewTmuxCollector_DefaultExecutor (0.00s)
=== RUN   TestListPanes_GenericError
--- PASS: TestListPanes_GenericError (0.00s)
=== RUN   TestBuildTmuxArgs_SocketPrecedence
=== RUN   TestBuildTmuxArgs_SocketPrecedence/CLI_参数优先级高于配置
=== RUN   TestBuildTmuxArgs_SocketPrecedence/配置指定的_SocketPath_转换为_-S
--- PASS: TestBuildTmuxArgs_SocketPrecedence (0.00s)
=== RUN   TestListPanes_Success
--- PASS: TestListPanes_Success (0.00s)
=== RUN   TestListPanes_ServerUnavailable
--- PASS: TestListPanes_ServerUnavailable (0.00s)
PASS
coverage: 93.1% of statements
ok  	tmux-ui/internal/collector	0.810s	coverage: 93.1% of statements
```
