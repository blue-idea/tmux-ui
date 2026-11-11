# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v2/ac/TASK-014-AC.md`
> 适用版本：v2
> 任务编号：TASK-014
> 执行日期：2026-11-12
> 执行人：Antigravity Agent

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-014 | v2.REQ-001-AC-001 | Unit | 成功监听 `StateStore` 状态跃迁事件，自动从最新 PaneDetail 提取 ANSI 文本与预览行，生成 `state_transition` 类型快照并落盘 | PASS | `go test -v ./internal/history/... -run TestEngine_StateTransitionCapturesSnapshot` | — |
| TASK-014 | v2.REQ-001-AC-003 | Unit | 周期性 Checkpoint 定时器自动对比输出哈希差异，当有新内容生成但未跃迁状态时落盘 `checkpoint` 保底快照；内存清理销毁 Pane 时已落盘快照安全留存 | PASS | `go test -v ./internal/history/... -run "TestEngine_PeriodicCheckpoint|TestEngine_DeadPanePreserved"` | — |

---

## 真实测试执行输出

```
=== RUN   TestEngine_StateTransitionCapturesSnapshot
--- PASS: TestEngine_StateTransitionCapturesSnapshot (0.10s)
=== RUN   TestEngine_PeriodicCheckpoint
--- PASS: TestEngine_PeriodicCheckpoint (0.10s)
=== RUN   TestEngine_DeadPanePreserved
--- PASS: TestEngine_DeadPanePreserved (0.10s)
PASS
coverage: 80.1% of statements
ok  	tmux-ui/internal/history	0.842s	coverage: 80.1% of statements
```
