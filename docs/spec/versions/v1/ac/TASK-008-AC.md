# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v1/ac/TASK-008-AC.md`
> 适用版本：v1
> 任务编号：TASK-008
> 执行日期：2026-11-11
> 执行人：Antigravity Agent

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-008 | v1.REQ-006-AC-001 | Web / Hook | 完成 `web/` Vite + React 18 + TypeScript + Tailwind CSS 工程初始化；实现领域 TypeScript 模型与 `useTmuxWS` Hook，通过 WebSocket 接收 snapshot 全量同步与 state_changed 增量状态更新；具备断线自动指数退避重连机制 | PASS | `src/hooks/useTmuxWS.test.ts` (4 passed) & `pnpm build` | — |

---

## 真实测试执行输出

```
> tmux-ui-web@1.0.0 test E:\Nextcloud\coding\agents\tmux-ui\web
> vitest run

 RUN  v2.1.9 E:/Nextcloud/coding/agents/tmux-ui/web

 ✓ src/hooks/useTmuxWS.test.ts (4 tests) 18ms

 Test Files  1 passed (1)
      Tests  4 passed (4)
   Start at  22:03:00
   Duration  1.59s (transform 39ms, setup 145ms, collect 59ms, tests 18ms, environment 418ms, prepare 72ms)
```

## 前端构建输出验证

```
> tmux-ui-web@1.0.0 build E:\Nextcloud\coding\agents\tmux-ui\web
> tsc && vite build

vite v6.4.6 building for production...
transforming...
✓ 1578 modules transformed.
rendering chunks...
computing gzip size...
dist/index.html                   0.53 kB │ gzip:  0.35 kB
dist/assets/index-Bd2ybcJ8.css    9.45 kB │ gzip:  2.60 kB
dist/assets/index-D5xtimqz.js   151.14 kB │ gzip: 48.58 kB
✓ built in 1.88s
```
