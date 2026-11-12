# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v3/ac/TASK-024-AC.md`
> 版本：v3
> 任务编号：TASK-024
> 执行日期：2026-11-12
> 执行人：Antigravity Agent (hns-sdd)

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-024 | v3.REQ-003-AC-001 | Unit | 前端单元测试验证 useTmuxWS 在未传 options.token 时自动从 tokenManager 解析有效 Token 并注入握手 URL ?token=... | PASS | `pnpm test src/hooks/useTmuxWS.test.ts` | — |

