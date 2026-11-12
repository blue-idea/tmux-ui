# 实施计划（Tasks）

> 文件路径：`docs/spec/versions/v3/tasks.md`
> 适用版本：v3 · 日期：2026-11-12
> 前置依据：`docs/spec/versions/v3/requirements.md` 与 `docs/spec/versions/v3/design.md`

执行时须严格遵循需求文档中的验收标准。每项任务通过 `_需求:` 引用需求编号，通过 `_测试类型:` 关联 QA 执行策略。

---

## 任务清单

---

- [x] **TASK-020 · Cloudflare Tunnel 部署模板、实战文档与反向代理适配**

  > 依赖：无

  - [x] 编写 Cloudflare Tunnel 生产与快速临时隧道实操部署指南 `docs/cloudflare-tunnel.md`
  - [x] 提供标准配置文件样例 `config/tunnel.yml.example`，配置 WebSocket 入口规则
  - [x] 优化 `internal/server/http_server.go`，增强 RealIP 中间件支持 `CF-Connecting-IP` 优先提取
  - [x] 编写 Go 测试用例验证反向代理请求头提取

  **验证方式：**
  ```bash
  go test -v ./internal/server -run TestHTTPServer_RealIP
  ```

  **验收证据：** Go 单元测试通过，`CF-Connecting-IP` 正确解析为远程客户端 IP，文档与配置样例完备。

  _需求: `requirements.md` 中 v3.REQ-001
  验收标准：v3.REQ-001-AC-002, v3.REQ-001-AC-004
  _测试类型: API

---

- [x] **TASK-021 · 服务端公网隧道集成、安全随机 Token 生成与直登 URL 打印**

  > 依赖：TASK-020

  - [x] 新建 `internal/tunnel` 模块，使用 `crypto/rand` 实现 32 字节高强度随机 Token 生成
  - [x] 实现公网直达免密 URL 格式化与控制台高亮卡片打印
  - [x] 扩展 `cmd/tmux-ui/main.go`，支持 `--tunnel`、`--public-url`、`--cloudflared-bin` 标志与对应环境变量
  - [x] 当公网模式启用且未指定 token 时，自动激活安全 Token 并输出防裸奔警告
  - [x] 编写 Go 单元测试覆盖 Token 生成、URL 拼装与安全防护拦截

  **验证方式：**
  ```bash
  go test -v ./internal/tunnel/...
  ```

  **验收证据：** 单元测试通过，终端启动时输出格式如 `https://.../?token=...` 的直登 URL。

  _需求: `requirements.md` 中 v3.REQ-001
  验收标准：v3.REQ-001-AC-001, v3.REQ-001-AC-003
  _测试类型: Unit

---

- [x] **TASK-022 · 前端 URL Token 解析、sessionStorage 记忆与地址栏隐私净化**

  > 依赖：无

  - [x] 新建 `web/src/auth/tokenManager.ts`，解析 URL 中的 `token` 或 `auth_token` 参数
  - [x] 实现 `sessionStorage` 会话持久化与优雅降级回退机制
  - [x] 实现 `scrubUrlToken()`，通过 `window.history.replaceState` 清除浏览器地址栏上的敏感 Token
  - [x] 编写 Vitest 单元测试覆盖各类 URL 参数、刷新回退与地址栏净化场景

  **验证方式：**
  ```bash
  cd web && pnpm test src/auth/tokenManager.test.ts
  ```

  **验收证据：** Vitest 测试全绿，Token 提取及地址栏无刷新净化完全符合预期。

  _需求: `requirements.md` 中 v3.REQ-002
  验收标准：v3.REQ-002-AC-001, v3.REQ-002-AC-002, v3.REQ-002-AC-003
  _测试类型: Unit

---

- [ ] **TASK-023 · 前端统一 HTTP API 客户端封装与全链路鉴权请求改造**

  > 依赖：TASK-022

  - [ ] 新建 `web/src/api/client.ts`，封装统一 `fetchWithAuth(url, options)` 客户端
  - [ ] 自动从 `tokenManager` 获取有效 Token 并注入 `Authorization: Bearer <token>`
  - [ ] 改造 `web/src/App.tsx` 中的 Pane 详情请求，使用 `fetchWithAuth` 替代原生 `fetch`
  - [ ] 改造 `web/src/hooks/useHistory.ts` 和 `web/src/components/PlaybackBar.tsx` 中的快照查询请求
  - [ ] 编写 Vitest 单元测试验证 Headers 注入与 401 拦截处理

  **验证方式：**
  ```bash
  cd web && pnpm test src/api/client.test.ts
  ```

  **验收证据：** Vitest 单元测试通过，所有 REST API 请求均自动附带 Bearer Token。

  _需求: `requirements.md` 中 v3.REQ-003
  验收标准：v3.REQ-003-AC-002
  _测试类型: Unit

---

- [ ] **TASK-024 · 前端 WebSocket 握手 Token 自动链式注入与断线重连**

  > 依赖：TASK-022

  - [ ] 改造 `web/src/hooks/useTmuxWS.ts`，自动从 `tokenManager` 解析当前有效 Token
  - [ ] 在 `buildWSUrl` 中自动将 Token 注入为 `?token=<token>` 参数
  - [ ] 保证断线自动重连时持续复用最新 Token 凭证
  - [ ] 编写与更新 Vitest 单元测试验证各种传参模式下的握手 URL

  **验证方式：**
  ```bash
  cd web && pnpm test src/hooks/useTmuxWS.test.ts
  ```

  **验收证据：** Vitest 测试通过，WebSocket URL 自动附带 `?token=...`。

  _需求: `requirements.md` 中 v3.REQ-003
  验收标准：v3.REQ-003-AC-001
  _测试类型: Unit

---

- [ ] **TASK-025 · 鉴权异常横幅组件与全链路集成端到端验收**

  > 依赖：TASK-021, TASK-023, TASK-024

  - [ ] 新建 `web/src/components/AuthErrorBanner.tsx`，展示优雅的英文鉴权失败与登录引导
  - [ ] 整合进 `web/src/App.tsx` 全局状态联动
  - [ ] 执行全套自动化测试（Go 后端测试 + 前端全部 Vitest 测试）
  - [ ] 编译单二进制可执行文件 `bin/tmux-ui` 并进行多视口集成验证与留证

  **验证方式：**
  ```bash
  go test ./... && cd web && pnpm test
  ```

  **验收证据：** 全部前后端单元测试通过，单二进制构建成功，生成 v3 验收测试报告。

  _需求: `requirements.md` 中 v3.REQ-003
  验收标准：v3.REQ-003-AC-003
  _测试类型: Unit
