# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v1/ac/TASK-005-AC.md`
> 适用版本：v1
> 任务编号：TASK-005
> 执行日期：2026-11-11
> 执行人：Antigravity Agent

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-005 | v1.REQ-003-AC-001 | Unit | `pane_dead == 1` 时最高优先级判定为 `exited`，覆盖率为 94.4% | PASS | `TestDetector_PriorityDeadExited` | — |
| TASK-005 | v1.REQ-003-AC-002 | Unit | 尾部文本匹配 `Do you want to proceed?` 或 `(y/n)` 等规则时准确判定为 `waiting_input` | PASS | `TestDetector_WaitingInputRegex` | — |
| TASK-005 | v1.REQ-003-AC-003 | Unit | 匹配 `working` 正则或终端输出哈希变动在超时窗口内时准确判定为 `working` | PASS | `TestDetector_WorkingEvaluation` | — |
| TASK-005 | v1.REQ-003-AC-004 | Unit | 进程为 bash/zsh 等纯 shell 且无 agent 子进程时判定为 `idle` | PASS | `TestDetector_ShellIdle` | — |
| TASK-005 | v1.REQ-003-AC-005 | Unit | 状态迁移时自动记录 TimelineEvent，环形缓冲严格限制 100 条上限（覆盖率 98.5%） | PASS | `TestStateStore_TimelineAppendOnStateTransition` & `TestStateStore_TimelineRingBufferBounded` | — |

---

## 真实测试执行输出

```
=== RUN   TestDetector_PriorityDeadExited
--- PASS: TestDetector_PriorityDeadExited (0.00s)
=== RUN   TestDetector_WaitingInputRegex
--- PASS: TestDetector_WaitingInputRegex (0.00s)
=== RUN   TestDetector_WorkingEvaluation
--- PASS: TestDetector_WorkingEvaluation (0.00s)
=== RUN   TestDetector_ShellIdle
--- PASS: TestDetector_ShellIdle (0.00s)
=== RUN   TestDetector_ErrorRegex
--- PASS: TestDetector_ErrorRegex (0.00s)
=== RUN   TestDetector_EdgeCases
--- PASS: TestDetector_EdgeCases (0.00s)
PASS
coverage: 94.4% of statements
ok  	tmux-ui/internal/detector	0.762s	coverage: 94.4% of statements

=== RUN   TestStateStore_TimelineAppendOnStateTransition
--- PASS: TestStateStore_TimelineAppendOnStateTransition (0.00s)
=== RUN   TestStateStore_TimelineRingBufferBounded
--- PASS: TestStateStore_TimelineRingBufferBounded (0.00s)
=== RUN   TestStateStore_EventBroadcasting
--- PASS: TestStateStore_EventBroadcasting (0.00s)
=== RUN   TestStateStore_CleanupDeadPanes
--- PASS: TestStateStore_CleanupDeadPanes (0.00s)
=== RUN   TestStateStore_ConcurrentSafety
--- PASS: TestStateStore_ConcurrentSafety (0.00s)
PASS
coverage: 98.5% of statements
ok  	tmux-ui/internal/store	1.534s	coverage: 98.5% of statements
```
