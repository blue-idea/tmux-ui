# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v1/ac/TASK-007-AC.md`
> 适用版本：v1
> 任务编号：TASK-007
> 执行日期：2026-11-11
> 执行人：Antigravity Agent

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-007 | v1.REQ-005-AC-003 | API / WS | 客户端连接 `/ws` 成功后立即接收到 `snapshot` 全量快照帧；当 `StateStore` 状态发生迁移时，Hub 自动向所有连接广播 `state_changed` 增量推送帧；支持 Ping/Pong 链路保活 | PASS | `TestWebSocket_SnapshotOnConnect`, `TestWebSocket_StateChangedBroadcast`, `TestWebSocket_PingPong` | — |
| TASK-007 | v1.REQ-005-AC-004 | API / WS | 配置 Token 场景下，未携带 Token 的 `/ws` 握手请求被拦截并返回 HTTP 401；携带有效 `?token=...` 查询参数的连接请求成功升级为 WebSocket | PASS | `TestWebSocket_TokenAuth` | — |

---

## 真实测试执行输出

```
=== RUN   TestWebSocket_SnapshotOnConnect
--- PASS: TestWebSocket_SnapshotOnConnect (0.00s)
=== RUN   TestWebSocket_StateChangedBroadcast
--- PASS: TestWebSocket_StateChangedBroadcast (0.00s)
=== RUN   TestWebSocket_TokenAuth
=== RUN   TestWebSocket_TokenAuth/无_Token_连接失败返回_401
=== RUN   TestWebSocket_TokenAuth/携带合法_Token_连接成功
--- PASS: TestWebSocket_TokenAuth (0.00s)
    --- PASS: TestWebSocket_TokenAuth/无_Token_连接失败返回_401 (0.00s)
    --- PASS: TestWebSocket_TokenAuth/携带合法_Token_连接成功 (0.00s)
=== RUN   TestWebSocket_PingPong
--- PASS: TestWebSocket_PingPong (0.00s)
PASS
coverage: 88.3% of statements
ok  	tmux-ui/internal/server	0.740s	coverage: 88.3% of statements
```

