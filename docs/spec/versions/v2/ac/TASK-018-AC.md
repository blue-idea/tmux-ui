# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v2/ac/TASK-018-AC.md`
> 适用版本：v2
> 任务编号：TASK-018
> 执行日期：2026-11-12
> 执行人：Antigravity Agent

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-018 | v2.REQ-004-AC-001 | 前端单元/E2E | `PlaybackBar` 组件渲染时间轴节点、状态色彩圆点及快照进度条；点击历史节点切换到快照并拉取快照 ANSI 输出填充终端 | PASS | `PlaybackBar.test.tsx` + `useHistory.test.ts` + Playwright 截图 | — |
| TASK-018 | v2.REQ-004-AC-002 | 前端单元/E2E | 在 Historical 回放模式下显示 "Historical Playback (时间)" 徽章与 "Return to Live" 按钮；点击恢复实时推流，并在新快照生成时保持同步 | PASS | `TerminalViewer.test.tsx` + `useHistory.test.ts` | — |

---

## 视觉回归测试留证（Visual Evidence）

- **时间轴历史回放播放器留证**：`docs/spec/versions/v2/ac/screenshots/TASK-018-playback.png`

---

## 真实测试执行输出

```
Test Files  9 passed (9)
     Tests  26 passed (26)
  Duration  6.62s

 ✓ src/hooks/useTmuxWS.test.ts (4)
 ✓ src/hooks/useHistory.test.ts (3)
 ✓ src/hooks/useNotifications.test.ts (3)
 ✓ src/hooks/useMediaQuery.test.ts (2)
 ✓ src/components/BottomSheet.test.tsx (2)
 ✓ src/components/OverviewGrid.test.tsx (4)
 ✓ src/components/PlaybackBar.test.tsx (2)
 ✓ src/components/TerminalViewer.test.tsx (3)
 ✓ src/components/TimelineDrawer.test.tsx (3)
```
