# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v1/ac/TASK-011-AC.md`
> 适用版本：v1
> 任务编号：TASK-011
> 执行日期：2026-11-11
> 执行人：Antigravity Agent

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-011 | v1.REQ-007-AC-002 | Web / UI | 实现 `TimelineDrawer` 抽屉组件，按时间正序展示状态迁移节点（时间戳、迁移前后的状态指示徽标、匹配规则、摘要说明）；支持抽屉展开、关闭与空状态提示 | PASS | `src/components/TimelineDrawer.test.tsx` (3 passed) & Playwright 视觉回归截图 | — |
| TASK-011 | v1.REQ-006-AC-003 | Web / Hook | 实现 `useNotifications` Hook，支持浏览器桌面通知权限请求；在智能体状态变迁为 `waiting_input` 或 `error` 时准确触发系统级桌面通知，其它状态静默 | PASS | `src/hooks/useNotifications.test.ts` (3 passed) | — |

---

## 真实测试执行输出

```
> tmux-ui-web@1.0.0 test E:\Nextcloud\coding\agents\tmux-ui\web
> vitest run

 RUN  v2.1.9 E:/Nextcloud/coding/agents/tmux-ui/web

 ✓ src/hooks/useTmuxWS.test.ts (4 tests) 21ms
 ✓ src/hooks/useNotifications.test.ts (3 tests) 16ms
 ✓ src/components/TerminalViewer.test.tsx (3 tests) 53ms
 ✓ src/components/TimelineDrawer.test.tsx (3 tests) 109ms
 ✓ src/components/OverviewGrid.test.tsx (4 tests) 143ms

 Test Files  5 passed (5)
      Tests  17 passed (17)
   Start at  22:12:49
   Duration  2.07s (transform 275ms, setup 870ms, collect 1.03s, tests 343ms, environment 2.79s, prepare 1.73s)
```

## 视觉回归测试（Visual Regression Screenshot）

- **截屏工具**：Playwright MCP (`playwright_navigate` + `playwright_screenshot`)
- **证据文件**：`docs/spec/versions/v1/ac/screenshots/TASK-011-timeline.png`
- **页面视图**：包含导航栏通知权限开关指示、时间线抽屉面板与状态跃迁徽标。

