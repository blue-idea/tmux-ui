# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v3/ac/TASK-026-AC.md`
> 版本：v3
> 任务编号：TASK-026
> 执行日期：2026-11-12
> 执行人：Antigravity Agent (hns-sdd)

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-026 | v3.REQ-004-AC-001 | Unit | TestFormatStatusTable 验证 ASCII 表格输出对齐、包含完整表头 (PANE ID, SESSION, AGENT, STATE, COMMAND, IDLE, PREVIEW) 及状态标签 | PASS | `go test -v ./internal/cli -run TestFormatStatusTable` | — |
| TASK-026 | v3.REQ-004-AC-002 | Unit | TestFormatStatusJSON 验证 --json 输出严格符合 []PaneSummary 结构，空列表输出 [] | PASS | `go test -v ./internal/cli -run TestFormatStatusJSON` | — |
| TASK-026 | v3.REQ-004-AC-003 | Unit | TestFetchStatusFromServer 验证 --server 支持远程拉取运行中的服务并附带 Bearer Token 鉴权 | PASS | `go test -v ./internal/cli -run TestFetchStatusFromServer` | — |
