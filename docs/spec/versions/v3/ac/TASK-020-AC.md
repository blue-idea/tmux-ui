# AC 验收矩阵（Acceptance Criteria Matrix）

> 文件路径：`docs/spec/versions/v3/ac/TASK-020-AC.md`
> 版本：v3
> 任务编号：TASK-020
> 执行日期：2026-11-12
> 执行人：Antigravity Agent (hns-sdd)

---

## 验收结果

| TASK ID | AC ID | QA 类型 | 实际结果摘要 | 状态 | 证据 | 错误详情 |
|---------|-------|:-------:|------------|:----:|------|---------|
| TASK-020 | v3.REQ-001-AC-002 | API | Go 单元测试 TestHTTPServer_RealIP_Cloudflare 验证 CF-Connecting-IP 与 X-Forwarded-For 正确解析 | PASS | `go test -v ./internal/server -run TestHTTPServer_RealIP_Cloudflare` | — |
| TASK-020 | v3.REQ-001-AC-004 | Manual | 编写生产与快速隧道实战部署文档 docs/cloudflare-tunnel.md 与配置文件样例 config/tunnel.yml.example | PASS | `config/tunnel.yml.example`, `docs/cloudflare-tunnel.md` | — |

