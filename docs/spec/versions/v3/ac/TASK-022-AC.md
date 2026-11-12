# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v3/ac/TASK-022-AC.md`
> 版本：v3
> 任务编号：TASK-022
> 执行日期：2026-11-12
> 执行人：Antigravity Agent (hns-sdd)

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-022 | v3.REQ-002-AC-001 | Unit | 前端单元测试验证 initAuthToken 从 ?token= 或 ?auth_token= 提取凭证并暂存至 sessionStorage | PASS | `pnpm test src/auth/tokenManager.test.ts` | — |
| TASK-022 | v3.REQ-002-AC-002 | Unit | 单元测试验证 scrubUrlToken 调用 history.replaceState 平滑擦除敏感 token 并保留其他参数 | PASS | `pnpm test src/auth/tokenManager.test.ts` | — |
| TASK-022 | v3.REQ-002-AC-003 | Unit | 单元测试验证 URL 无参数时无缝回退读取 sessionStorage 缓存的 Token | PASS | `pnpm test src/auth/tokenManager.test.ts` | — |

