# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v3/ac/TASK-025-AC.md`
> 版本：v3
> 任务编号：TASK-025
> 执行日期：2026-11-12
> 执行人：Antigravity Agent (hns-sdd)

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-025 | v3.REQ-003-AC-003 | Unit | AuthErrorBanner 单元与组件渲染测试通过，捕获 401 时展示标准英文告警文案及 Dismiss 按钮 | PASS | `pnpm test src/components/AuthErrorBanner.test.tsx` | — |
| TASK-025 | v3.REQ-002-AC-001 | E2E | Playwright 实测免密 URL 直登模式 (Desktop 1280x800)：提取 Token 自动握手进入 connected 状态 | PASS | `docs/spec/versions/v3/ac/screenshots/TASK-025-desktop-url-token-login.png` | — |
| TASK-025 | v3.REQ-002-AC-002 | E2E | Playwright 实测移动视口 (375x667 iPhone/Telegram Mini App)：自适应渲染，地址栏参数静默净化 | PASS | `docs/spec/versions/v3/ac/screenshots/TASK-025-mobile-url-token-login.png` | — |

