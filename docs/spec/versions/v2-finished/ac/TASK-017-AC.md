# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v2/ac/TASK-017-AC.md`
> 适用版本：v2
> 任务编号：TASK-017
> 执行日期：2026-11-12
> 执行人：Antigravity Agent

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-017 | v2.REQ-005-AC-001 | E2E | 移动端视口（iPhone 14，390px 宽）下打开时间线抽屉时，组件自适应为自下而上滑出的 `BottomSheet`（最大高度 85vh），包含顶部手势把手，支持点击遮罩或右上角按钮关闭 | PASS | Vitest 单元测试与 Playwright 移动端视觉截屏 | — |
| TASK-017 | v2.REQ-005-AC-002 | E2E | 移动端终端查看器字体与容器自适应缩小，右下角新增浮动快速回到底部按钮（FAB），所有可触控操作按钮尺寸均满足 $\ge 44 \times 44\text{px}$ 规范 | PASS | DOM 检查与 Playwright 截屏留证 | — |

---

## 视觉回归测试留证（Visual Evidence）

- **移动端视口（iPhone 14，390x664）**：`docs/spec/versions/v2/ac/screenshots/TASK-017-mobile-overview.png`

---

## 真实测试执行输出

```
 ✓ src/components/BottomSheet.test.tsx (2 tests)
 ✓ src/components/TimelineDrawer.test.tsx (3 tests)
 ✓ src/components/TerminalViewer.test.tsx (3 tests)
 ✓ src/hooks/useMediaQuery.test.ts (2 tests)
 ✓ src/components/OverviewGrid.test.tsx (4 tests)
 ✓ src/hooks/useNotifications.test.ts (3 tests)
 ✓ src/hooks/useTmuxWS.test.ts (4 tests)

 Test Files  7 passed (7)
      Tests  21 passed (21)
```
