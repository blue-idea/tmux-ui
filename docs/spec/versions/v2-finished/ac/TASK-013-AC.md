# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v2/ac/TASK-013-AC.md`
> 适用版本：v2
> 任务编号：TASK-013
> 执行日期：2026-11-12
> 执行人：Antigravity Agent

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-013 | v2.REQ-001-AC-001 | Unit | 成功创建 `SnapshotRecord` 并追加写入本地 `.jsonl` 文件，包含有效 SnapshotID、时间戳、PaneID 与 ANSI 内容，且可通过 GetSnapshot 完整还原 | PASS | `go test -v ./internal/history/... -run TestStorage_AppendAndReadRecord` | — |
| TASK-013 | v2.REQ-001-AC-002 | Unit | 写入超出 `max_snapshots_per_pane` 阈值时，自动触发 FIFO 原子截断并保留最新的 N 条记录 | PASS | `go test -v ./internal/history/... -run TestStorage_FIFOPruning` | — |
| TASK-013 | v2.REQ-001-AC-003 | Unit | 目录安全转义 `SanitizePaneID` 防止路径穿越，多协程高并发写入零数据竞争零锁死 | PASS | `go test -v ./internal/history/... -run "TestStorage_SanitizePaneID|TestStorage_ConcurrentSafety"` | — |

---

## 真实测试执行输出

```
=== RUN   TestStorage_SanitizePaneID
=== RUN   TestStorage_SanitizePaneID/%0
=== RUN   TestStorage_SanitizePaneID/%15
=== RUN   TestStorage_SanitizePaneID/../etc/passwd
=== RUN   TestStorage_SanitizePaneID/pane/with/slash
=== RUN   TestStorage_SanitizePaneID/normal-id
--- PASS: TestStorage_SanitizePaneID (0.00s)
=== RUN   TestStorage_AppendAndReadRecord
--- PASS: TestStorage_AppendAndReadRecord (0.00s)
=== RUN   TestStorage_QueryMetadataList
--- PASS: TestStorage_QueryMetadataList (0.00s)
=== RUN   TestStorage_FIFOPruning
--- PASS: TestStorage_FIFOPruning (0.01s)
=== RUN   TestStorage_ConcurrentSafety
--- PASS: TestStorage_ConcurrentSafety (0.02s)
=== RUN   TestStorage_EdgeCasesAndErrors
--- PASS: TestStorage_EdgeCasesAndErrors (0.00s)
PASS
coverage: 80.6% of statements
ok  	tmux-ui/internal/history	0.531s	coverage: 80.6% of statements
```
