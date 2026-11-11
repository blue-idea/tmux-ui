# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v1/ac/TASK-006-AC.md`
> 适用版本：v1
> 任务编号：TASK-006
> 执行日期：2026-11-11
> 执行人：Antigravity Agent

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-006 | v1.REQ-005-AC-001 | API | `GET /api/v1/panes` 返回 HTTP 200，支持 `status` 与 `session` 过滤；`GET /api/v1/health` 及 `/api/v1/config/agents` 返回符合 Schema | PASS | `TestHTTP_HealthAndConfig` & `TestHTTP_ListPanes` | — |
| TASK-006 | v1.REQ-005-AC-002 | API | `GET /api/v1/panes/{id}` 支持裸数字 `1` 规整为 `%1` 与 `%251` 解码；不存在时返回 HTTP 404 及 `PANE_NOT_FOUND` 标准错误 JSON | PASS | `TestHTTP_GetPaneDetail_FlexibleRouting` | — |
| TASK-006 | v1.REQ-005-AC-004 | API | 配置 Token 时无 Token 或无效 Token 请求均被拦截返回 HTTP 401 及 `UNAUTHORIZED`；携带合法 Header/Query Token 正常放行 | PASS | `TestHTTP_AuthMiddleware` | — |

---

## 真实测试执行输出

```
=== RUN   TestHTTP_HealthAndConfig
=== RUN   TestHTTP_HealthAndConfig/GET_/api/v1/health_返回系统状态
=== RUN   TestHTTP_HealthAndConfig/GET_/api/v1/config/agents_返回智能体规则列表
--- PASS: TestHTTP_HealthAndConfig (0.00s)
=== RUN   TestHTTP_ListPanes
=== RUN   TestHTTP_ListPanes/获取全部_Pane
=== RUN   TestHTTP_ListPanes/按状态过滤_?status=waiting_input
=== RUN   TestHTTP_ListPanes/按会话过滤_?session=prod
--- PASS: TestHTTP_ListPanes (0.01s)
=== RUN   TestHTTP_GetPaneDetail_FlexibleRouting
=== RUN   TestHTTP_GetPaneDetail_FlexibleRouting/支持裸数字_ID_/api/v1/panes/1
=== RUN   TestHTTP_GetPaneDetail_FlexibleRouting/支持编码字符_/api/v1/panes/%251
=== RUN   TestHTTP_GetPaneDetail_FlexibleRouting/不存在的_Pane_返回_404_及标准错误_JSON
--- PASS: TestHTTP_GetPaneDetail_FlexibleRouting (0.00s)
=== RUN   TestHTTP_AuthMiddleware
=== RUN   TestHTTP_AuthMiddleware/未携带_Token_返回_401
=== RUN   TestHTTP_AuthMiddleware/携带无效_Token_返回_401
=== RUN   TestHTTP_AuthMiddleware/携带有效_Header_Token_返回_200
=== RUN   TestHTTP_AuthMiddleware/携带有效_Query_Param_Token_返回_200
--- PASS: TestHTTP_AuthMiddleware (0.00s)
PASS
coverage: 94.6% of statements
ok  	tmux-ui/internal/server	0.908s	coverage: 94.6% of statements
```
