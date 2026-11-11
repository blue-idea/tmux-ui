# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v2/ac/TASK-016-AC.md`
> 适用版本：v2
> 任务编号：TASK-016
> 执行日期：2026-11-12
> 执行人：Antigravity Agent

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-016 | v2.REQ-004-AC-001 | E2E | 移动端视口（iPhone 14，390px 宽）下 Header 自动折叠次要文本标签，紧凑对齐，无横向溢出滚动条 | PASS | Playwright 视觉快照与 Vitest 测试 | — |
| TASK-016 | v2.REQ-004-AC-002 | E2E | FilterBar 在小屏幕下支持触摸横向平滑滚动（Touch Scroll），搜索输入框自适应 100% 容器全宽且触控高度满足标准 | PASS | Playwright 视觉快照与 DOM 样式检查 | — |
| TASK-016 | v2.REQ-004-AC-003 | E2E | OverviewGrid 与 PaneCard 在移动端自动切换为自适应单列卡片垂直排布，内边距与字体缩放优雅，终端输出预览不被截断 | PASS | Playwright 移动端视觉截屏留存 | — |

---

## 视觉回归测试留证（Visual Evidence）

- **桌面端（1280x800）**：`docs/spec/versions/v2/ac/screenshots/TASK-016-desktop-overview.png`
- **移动端（iPhone 14，390x664）**：`docs/spec/versions/v2/ac/screenshots/TASK-016-mobile-overview.png`

---

## 真实测试执行输出

```
 ✓ src/hooks/useMediaQuery.test.ts (2 tests)
 ✓ src/components/OverviewGrid.test.tsx (4 tests)
 ✓ src/components/TerminalViewer.test.tsx (3 tests)
 ✓ src/components/TimelineDrawer.test.tsx (3 tests)
 ✓ src/hooks/useNotifications.test.ts (3 tests)
 ✓ src/hooks/useTmuxWS.test.ts (4 tests)

 Test Files  6 passed (6)
      Tests  19 passed (19)
```
