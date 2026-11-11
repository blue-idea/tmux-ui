# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v1/ac/TASK-010-AC.md`
> 适用版本：v1
> 任务编号：TASK-010
> 执行日期：2026-11-11
> 执行人：Antigravity Agent

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-010 | v1.REQ-007-AC-001 | Web / UI | 实现 `TerminalViewer` 组件，正确集成 `@xterm/xterm` 与 `@xterm/addon-fit`，支持 ANSI 色彩字符流保真渲染与容器自适应尺寸调整 | PASS | `src/components/TerminalViewer.test.tsx` (3 passed) | — |
| TASK-010 | v1.REQ-007-AC-003 | Web / UI | 严格只读保护模式启用（`disableStdin: true`，页面常驻 Read-Only Mode 标识），提供便捷 ANSI 内容一键复制，展示关联进程树表格（PID/PPID/Command）与元数据 | PASS | `src/components/TerminalViewer.test.tsx` & Playwright 视觉回归截图 | — |

---

## 真实测试执行输出

```
> tmux-ui-web@1.0.0 test E:\Nextcloud\coding\agents\tmux-ui\web
> vitest run

 RUN  v2.1.9 E:/Nextcloud/coding/agents/tmux-ui/web

 ✓ src/hooks/useTmuxWS.test.ts (4 tests) 24ms
 ✓ src/components/TerminalViewer.test.tsx (3 tests) 57ms
 ✓ src/components/OverviewGrid.test.tsx (4 tests) 121ms

 Test Files  3 passed (3)
      Tests  11 passed (11)
   Start at  22:09:36
   Duration  1.89s (transform 370ms, setup 481ms, collect 632ms, tests 202ms, environment 1.32s, prepare 277ms)
```

## 视觉回归测试（Visual Regression Screenshot）

- **截屏工具**：Playwright MCP (`playwright_navigate` + `playwright_screenshot`)
- **证据文件**：`docs/spec/versions/v1/ac/screenshots/TASK-010-terminal.png`
- **页面视图**：包含返回总览导航、Pane 标题与状态灯、只读安全徽标、一键复制、xterm.js 暗黑终端视口与侧边进程树面板。
