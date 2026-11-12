# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v3/ac/TASK-027-AC.md`
> 版本：v3
> 任务编号：TASK-027
> 执行日期：2026-11-12
> 执行人：Antigravity Agent (hns-sdd)

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-027 | v3.REQ-004-AC-001 | Manual | 根目录 README.md 详尽列出 tmux-ui status / ps / ls 命令行使用说明、参数列表与对齐表格展示 | PASS | `README.md` 中 `## 💻 命令行实时状态查询 (CLI Mode)` | — |
| TASK-027 | v3.REQ-004-AC-002 | CLI | 真实执行 bin/tmux-ui.exe status 与 bin/tmux-ui.exe status --json，终端输出完全符合规约 | PASS | 真实命令执行输出日志 | — |
