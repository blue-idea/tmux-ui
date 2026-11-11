# 锁定技术栈（Tech Stack）

> 文件路径：`docs/spec/steering/tech-stack.md`
> 创建步骤：STEP 1（首次立规矩时生成）
> ⚠️ 本文件**不随 version 变更**。技术栈变更属于重大决策，应升级 version 并记录原因

---

## 后端技术栈 (Go Backend)

- **语言**：Go (`go 1.22+` / `1.23+`)
- **二进制分发**：单二进制编译，采用 Go 1.16+ 原生 `embed` 打包前端静态资源
- **Web 服务与 API**：Go 标准库 `net/http` / 轻量路由（如 `go-chi/chi/v5`）
- **实时通信**：WebSocket（推送拓扑变更与 pane 输出 diff）
- **配置解析**：YAML 解析器（`gopkg.in/yaml.v3`），统一配置文件置于 `config/agents.yaml`
- **外部系统调用**：`os/exec` 调用本机 `tmux` CLI 与 `ps` 命令（严格处理 cross-platform 参数兼容）

---

## 前端技术栈 (React Frontend)

- **框架与构建工具**：React 18/19 + Vite + TypeScript
- **样式方案**：Tailwind CSS
- **终端渲染引擎**：`@xterm/xterm`（配合 `@xterm/addon-fit`、ANSI 颜色渲染）
- **包管理器**：**pnpm**（统一使用，严格版本锁定 `pnpm-lock.yaml`，禁止混用 npm/yarn）
- **TypeScript 规范**：
  - 启用严格模式（`strict: true`）
  - 严禁使用 `any` 或 `as any`，使用明确接口定义和泛型
  - 任务完成后自动执行 `pnpm exec tsc --noEmit` 检查

---

## 测试框架

| 测试类型 | 框架 / 工具 | 适用场景 | 验收要求 |
|---------|-----------|---------|---------|
| Go 单元测试 | `go test` + `testify` | Collector、Detector 启发式分析、State Store | 输出测试结果与覆盖率 |
| Go API / WS 集成测试 | `net/http/httptest` + `curl` | REST 接口、WebSocket 广播握手与推流 | 实际执行验证 |
| React 单元 / 组件测试 | Vitest + React Testing Library | 状态卡片、筛选栏、时间线渲染 | 覆盖主要交互路径 |
| 视觉回归与 E2E | Playwright MCP | UI 布局、状态指示灯、xterm 渲染截屏 | **必须保留截图与 Diff** |
| 性能基准测试 | `go test -bench` | 终端字符正则匹配与哈希 diff 计算 | 输出 Benchmark 耗时 |

---

## 安全与运行环境约束

- **网络绑定**：默认仅监听 `127.0.0.1`，支持启动生成随机 Token 认证。
- **只读原则**：MVP 阶段仅执行只读监控（`list-panes`, `capture-pane`）；写操作（如 `send-keys`）必须显式通过配置开关开启。
- **环境兼容**：
  - 支持 macOS（Darwin）与 Linux，检测 `tmux -V` 并动态适配 format 变量。
  - `ps` 查询进程树使用兼容语法 `ps -axo pid=,ppid=,command=`。

