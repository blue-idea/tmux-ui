# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v2/ac/TASK-015-AC.md`
> 适用版本：v2
> 任务编号：TASK-015
> 执行日期：2026-11-12
> 执行人：Antigravity Agent

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-015 | v2.REQ-002-AC-001 | API | 成功实现 `GET /api/v1/panes/{id}/snapshots`，支持 `since` / `limit` 过滤与裸数字宽容路由，返回轻量级时间轴元数据列表（HTTP 200） | PASS | `go test -v ./internal/server/... -run TestSnapshotAPI_ListSnapshots` | — |
| TASK-015 | v2.REQ-002-AC-002 | API | 成功实现 `GET /api/v1/panes/{id}/snapshots/{snapshot_id}`，完整还原指定时刻的 ANSI 终端内容与元数据（HTTP 200） | PASS | `go test -v ./internal/server/... -run TestSnapshotAPI_GetSnapshotDetail` | — |
| TASK-015 | v2.REQ-002-AC-003 | API | 访问不存在的快照 ID 或未初始化 Pane 返回标准 HTTP 404 JSON 错误信封（`SNAPSHOT_NOT_FOUND`）；扩展 WSHub 广播 `snapshot_created` 帧成功 | PASS | `go test -v ./internal/server/... -run "TestSnapshotAPI_GetSnapshotDetail|TestSnapshotAPI_WSHubBroadcast"` | — |

---

## 真实测试执行输出

```
=== RUN   TestSnapshotAPI_ListSnapshots
--- PASS: TestSnapshotAPI_ListSnapshots (0.00s)
=== RUN   TestSnapshotAPI_GetSnapshotDetail
--- PASS: TestSnapshotAPI_GetSnapshotDetail (0.00s)
=== RUN   TestSnapshotAPI_WSHubBroadcast
--- PASS: TestSnapshotAPI_WSHubBroadcast (0.05s)
PASS
coverage: 86.8% of statements
ok  	tmux-ui/internal/server	0.626s	coverage: 86.8% of statements
```
