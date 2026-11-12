# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v3/ac/TASK-021-AC.md`
> 版本：v3
> 任务编号：TASK-021
> 执行日期：2026-11-12
> 执行人：Antigravity Agent (hns-sdd)

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-021 | v3.REQ-001-AC-001 | Unit | Go 测试 TestFormatPublicURL 验证公网基础 URL 自动规范化追加 ?token=... 免密直登参数 | PASS | `go test -v ./internal/tunnel -run TestFormatPublicURL` | — |
| TASK-021 | v3.REQ-001-AC-003 | Unit | Go 测试 TestGenerateSecureToken 与 TestEnsureTunnelSecurity 验证公网模式未配 Token 时自动生成 32 字节高熵随机 Token 防裸奔 | PASS | `go test -v ./internal/tunnel -run TestEnsureTunnelSecurity` | — |
