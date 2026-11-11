# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v1/ac/TASK-003-AC.md`
> 适用版本：v1
> 任务编号：TASK-003
> 执行日期：2026-11-11
> 执行人：Antigravity Agent

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-003 | v1.REQ-001-AC-002 | Unit | 成功解析 `ps -axo pid=,ppid=,command=` 输出并构造 PID-PPID 树，通过根 PID 提取所有子孙进程并准确识别智能体类型（覆盖率为 94.8%） | PASS | `go test -v ./internal/collector/... -run TestProcessTree` | — |

---

## 真实测试执行输出

```
=== RUN   TestProcessTree_ParseRawOutput
--- PASS: TestProcessTree_ParseRawOutput (0.00s)
=== RUN   TestProcessTree_GetPaneProcesses
--- PASS: TestProcessTree_GetPaneProcesses (0.00s)
=== RUN   TestProcessTree_MatchAgentType
=== RUN   TestProcessTree_MatchAgentType/匹配到_claude-code
=== RUN   TestProcessTree_MatchAgentType/匹配到_aider
=== RUN   TestProcessTree_MatchAgentType/纯_shell_未匹配到智能体
--- PASS: TestProcessTree_MatchAgentType (0.00s)
=== RUN   TestProcessTree_ExecutorError
--- PASS: TestProcessTree_ExecutorError (0.00s)
=== RUN   TestProcessTree_EdgeCases
=== RUN   TestProcessTree_EdgeCases/默认构造函数返回可用实例
=== RUN   TestProcessTree_EdgeCases/畸形输出行安全忽略
--- PASS: TestProcessTree_EdgeCases (0.00s)
PASS
coverage: 94.8% of statements
ok  	tmux-ui/internal/collector	0.792s	coverage: 94.8% of statements
```
