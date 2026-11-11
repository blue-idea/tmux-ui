# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v1/ac/TASK-004-AC.md`
> 适用版本：v1
> 任务编号：TASK-004
> 执行日期：2026-11-11
> 执行人：Antigravity Agent

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-004 | v1.REQ-002-AC-001 | Unit | 成功执行 `capture-pane -p -e -t %id` 并且完整保留 ANSI 彩色转义字符 | PASS | `TestCapturePane_ANSIEscapeRetained` | — |
| TASK-004 | v1.REQ-002-AC-002 | Unit | 提取末尾 N 行计算 64 位内容哈希，内容变动时准确刷新 `LastActiveAt`，无变动时不刷新 | PASS | `TestPaneOutputTracker_HashAndActivityTracking` | — |
| TASK-004 | v1.REQ-002-AC-003 | Unit | 当 Pane 为已死（`isDead=true`）时标记 `IsFinalized=true` 固化缓存，停止后续无效轮询 | PASS | `TestPaneOutputTracker_DeadPaneFinalized` | — |

---

## 真实测试执行输出

```
=== RUN   TestCapturePane_ANSIEscapeRetained
--- PASS: TestCapturePane_ANSIEscapeRetained (0.00s)
=== RUN   TestExtractTailLines
--- PASS: TestExtractTailLines (0.00s)
=== RUN   TestPaneOutputTracker_HashAndActivityTracking
--- PASS: TestPaneOutputTracker_HashAndActivityTracking (0.00s)
=== RUN   TestPaneOutputTracker_DeadPaneFinalized
--- PASS: TestPaneOutputTracker_DeadPaneFinalized (0.00s)
=== RUN   TestCapturePane_Errors
--- PASS: TestCapturePane_Errors (0.00s)
PASS
coverage: 96.3% of statements
ok  	tmux-ui/internal/collector	0.809s	coverage: 96.3% of statements
```
