# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v3/ac/TASK-023-AC.md`
> 版本：v3
> 任务编号：TASK-023
> 执行日期：2026-11-12
> 执行人：Antigravity Agent (hns-sdd)

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-023 | v3.REQ-003-AC-002 | Unit | 前端单元测试验证 apiClient / fetchWithAuth 统一附加 Authorization: Bearer <token> 请求头并改造 App 与 useHistory | PASS | `pnpm test src/api/client.test.ts` | — |
