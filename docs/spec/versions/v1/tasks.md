# 实施计划（Tasks）

> 文件路径：`docs/spec/versions/v1/tasks.md`
> 适用版本：v1 · 生成日期：2026-11-11
> 前置依据：`docs/spec/versions/v1/requirements.md` 与 `docs/spec/versions/v1/design.md`

执行时须严格遵循需求文档与验收标准，严格遵循测试驱动开发（TDD）红-绿-重构循环。
每个任务必须执行真实命令验证并保留测试通过证据。

---

## 任务清单

---

- [x] **TASK-001 · 项目基础脚手架、配置模型与解析器**

  > 依赖：无

  - [x] 初始化 Go 模块（`go.mod`）与基本目录架构（`cmd/`, `internal/`, `config/`）
  - [x] 实现领域实体模型 `internal/model`（`State` 枚举、`PaneSummary`、`PaneDetail`、`TimelineEvent`、`AgentRule`）
  - [x] 编写 `internal/config` 加载并解析 `config/agents.yaml`，实现默认内置规则降级兜底

  **验证方式：**
  ```bash
  go test -v ./internal/config/... ./internal/model/...
  ```

  **验收证据：** 单元测试 100% 通过，能正确解析 Claude Code / Codex / Aider 配置，并在缺失配置文件时成功降级至默认规则。

  _需求: `requirements.md` 中 v1.REQ-004
  验收标准：v1.REQ-004-AC-001, v1.REQ-004-AC-002
  _测试类型: Unit

---

- [x] **TASK-002 · Tmux CLI 采集器与版本探测兼容**

  > 依赖：TASK-001

  - [x] 抽象 `TmuxExecutor` 接口以支持测试 Mock 注入
  - [x] 实现 `tmux -V` 版本检测与 format 格式自适应降级逻辑
  - [x] 实现 `tmux list-panes` 命令构建，支持命令行 `--socket-name` / `--socket-path` 及配置文件优先级覆盖
  - [x] 实现输出字符串解析器，安全捕获 tmux 未启动或 socket 不可达的 typed error（`ErrTmuxUnavailable`）

  **验证方式：**
  ```bash
  go test -v ./internal/collector/... -run TestTmuxCollector
  ```

  **验收证据：** 单元测试覆盖不同 tmux 版本输出样例、自定义 socket 参数拼接以及服务未启动异常处理。

  _需求: `requirements.md` 中 v1.REQ-001
  验收标准：v1.REQ-001-AC-001, v1.REQ-001-AC-003, v1.REQ-001-AC-004, v1.REQ-001-AC-005
  _测试类型: Unit

---

- [x] **TASK-003 · 跨平台进程树探测器与 Agent 进程关联**

  > 依赖：TASK-001, TASK-002

  - [x] 实现跨平台兼容命令 `ps -axo pid=,ppid=,command=` 执行与输出流解析器
  - [x] 实现从 PID/PPID 构造树状进程层级关系（`ProcessInfo`）算法
  - [x] 根据 Pane 根 PID，递归遍历提取子进程列表，并与 `AgentRule.MatchProcess` 进行规则匹配

  **验证方式：**
  ```bash
  go test -v ./internal/collector/... -run TestProcessTree
  ```

  **验收证据：** 单元测试验证 Linux 与 macOS 格式 ps 输出解析，正确识别 shell 下挂载的 agent 子进程。

  _需求: `requirements.md` 中 v1.REQ-001
  验收标准：v1.REQ-001-AC-002
  _测试类型: Unit

---

- [x] **TASK-004 · 终端 ANSI 输出捕获与哈希差异跟踪**

  > 依赖：TASK-002

  - [x] 实现 `tmux capture-pane -p -e -t %id` 命令执行与 ANSI 彩色转义字符完整保留
  - [x] 实现终端末尾 N 行文本提取与 64 位内容哈希算法
  - [x] 实现活跃状态追踪逻辑：哈希变动时刷新 `LastActiveAt`，超时无变动识别为可能 idle
  - [x] 对 `pane_dead == 1` 的 Pane 标记状态终结并停止无效轮询

  **验证方式：**
  ```bash
  go test -v ./internal/collector/... -run TestCapturePane
  ```

  **验收证据：** 单元测试通过，捕获输出 ANSI 码无损坏，哈希变动与活跃时间戳准确刷新。

  _需求: `requirements.md` 中 v1.REQ-002
  验收标准：v1.REQ-002-AC-001, v1.REQ-002-AC-002, v1.REQ-002-AC-003
  _测试类型: Unit

---

- [x] **TASK-005 · 多智能体状态判定引擎与内存存储**

  > 依赖：TASK-003, TASK-004

  - [x] 实现多智能体状态决策链：`pane_dead` 优先 -> `waiting_input` 正则 -> `error` 正则 -> `working` 正则 -> 输出哈希活跃度 -> 纯 shell `idle`
  - [x] 实现线程安全 `StateStore`（基于 `sync.RWMutex` 保护）
  - [x] 实现上限 100 条记录的环形缓冲（RingBuffer）时间线事件存储，支持 10 分钟 TTL 销毁 pane 清理

  **验证方式：**
  ```bash
  go test -v ./internal/detector/... ./internal/store/...
  ```

  **验收证据：** 单元测试覆盖全部状态迁移场景（Claude Code 提问、运行中、异常退出、空闲 shell）并验证并发安全。

  _需求: `requirements.md` 中 v1.REQ-003
  验收标准：v1.REQ-003-AC-001, v1.REQ-003-AC-002, v1.REQ-003-AC-003, v1.REQ-003-AC-004, v1.REQ-003-AC-005
  _测试类型: Unit

---

- [x] **TASK-006 · 后端 HTTP REST API 与 Token 认证中间件**

  > 依赖：TASK-005

  - [x] 基于 Chi 路由实现 `GET /api/v1/health` 与 `GET /api/v1/config/agents`
  - [x] 实现 `GET /api/v1/panes` 列表查询与 status/session 过滤
  - [x] 实现 `GET /api/v1/panes/{id}` 宽容路由参数解析（支持裸数字 `1` 自动规整为 `%1` 及 `%251` URL 解码）
  - [x] 实现可选 Bearer Token 鉴权中间件与统一标准 JSON 错误响应规范

  **验证方式：**
  ```bash
  go test -v ./internal/server/... -run TestHTTP
  ```

  **验收证据：** API 测试展示 200 / 401 / 404 真实返回值，验证宽容路由与 Token 拦截有效。

  _需求: `requirements.md` 中 v1.REQ-005
  验收标准：v1.REQ-005-AC-001, v1.REQ-005-AC-002, v1.REQ-005-AC-004
  _测试类型: API

---

- [x] **TASK-007 · WebSocket Hub 实时推流服务**

  > 依赖：TASK-005, TASK-006

  - [x] 实现 WebSocket 协议升级、Query Token 鉴权与 Client 连接管理 Hub
  - [x] 实现客户端连入后立即下发 `snapshot` 全量快照
  - [x] 接入 StateStore 变迁广播通道，向客户端推流 `pane_updated`、`state_changed` 与 `output_diff`
  - [x] 实现 `ping` / `pong` 心跳保活与异常断开清理机制

  **验证方式：**
  ```bash
  go test -v ./internal/server/... -run TestWebSocket
  ```

  **验收证据：** WebSocket 握手成功，正确收发 snapshot 帧，并在模拟状态变迁时接收到广播事件。

  _需求: `requirements.md` 中 v1.REQ-005
  验收标准：v1.REQ-005-AC-003
  _测试类型: API

---

- [x] **TASK-008 · 前端 Vite 脚手架与 WebSocket 通信 Hook**

  > 依赖：TASK-007

  - [x] 使用 Vite + React + Tailwind CSS + TypeScript 初始化 `web/` 工程并配置严格类型规范
  - [x] 声明前端核心 TypeScript 接口定义（`PaneSummary`, `PaneDetail`, `TimelineEvent`）
  - [x] 实现 `useTmuxWS` Hook（自动断线重连、快照同步、局部状态更新）

  **验证方式：**
  ```bash
  pnpm --prefix web exec tsc --noEmit && pnpm --prefix web test --run
  ```

  **验收证据：** 前端 TypeScript 编译 0 错误，单元测试验证 WebSocket 消息解析与状态归并逻辑。

  _需求: `requirements.md` 中 v1.REQ-006
  验收标准：v1.REQ-006-AC-001
  _测试类型: Unit

---

- [x] **TASK-009 · Web 总览卡片墙与多维状态筛选**

  > 依赖：TASK-008

  - [x] 开发 `OverviewGrid` 响应式网格容器组件
  - [x] 开发 `PaneCard` 组件：渲染 Session/Window 标题、Pane ID、Agent Badge、色彩编码状态指示灯、空闲时间与末尾 ANSI 预览
  - [x] 开发 `FilterBar` 状态快速过滤导航与各状态实时计数徽章

  **验证方式：**
  ```bash
  pnpm --prefix web test --run && playwright test overview
  ```

  **验收证据：** 组件单元测试通过，Playwright 截图生成总览卡片墙渲染证据及按状态筛选的前后对比截图。

  _需求: `requirements.md` 中 v1.REQ-006
  验收标准：v1.REQ-006-AC-001, v1.REQ-006-AC-002
  _测试类型: E2E

---

- [x] **TASK-010 · 终端详情页与 xterm.js ANSI 渲染器**

  > 依赖：TASK-008, TASK-009

  - [x] 安装并配置 `@xterm/xterm` 与 `@xterm/addon-fit`
  - [x] 开发 `TerminalViewer` 组件，挂载 xterm.js 实例并渲染全量 ANSI 彩色字符流
  - [x] 实现终端自适应容器大小（Resize / FitAddon）与严格只读保护（禁用键盘输入并提示 Read-Only）
  - [x] 配置前端路由支持 `/panes/:id` 导航与数据拉取

  **验证方式：**
  ```bash
  playwright test terminal
  ```

  **验收证据：** Playwright 截图留存 xterm.js 渲染彩色终端输出效果，验证无法键盘键入命令。

  _需求: `requirements.md` 中 v1.REQ-007
  验收标准：v1.REQ-007-AC-001, v1.REQ-007-AC-003
  _测试类型: E2E

---

- [x] **TASK-011 · 状态变迁时间线抽屉与浏览器交互通知**

  > 依赖：TASK-009, TASK-010

  - [x] 开发 `TimelineDrawer` 抽屉组件，按时间正序展示状态迁移节点（时间戳、迁移原因、规则匹配摘要）
  - [x] 实现 `useNotifications` Hook 请求浏览器桌面通知权限
  - [x] 当智能体状态跃迁为 `waiting_input` 或 `error` 时触发系统级桌面通知

  **验证方式：**
  ```bash
  pnpm --prefix web test --run && playwright test timeline_and_notify
  ```

  **验收证据：** 单元测试通过，Playwright 截图展示时间线抽屉与通知触发模拟验证。

  _需求: `requirements.md` 中 v1.REQ-006, v1.REQ-007
  验收标准：v1.REQ-006-AC-003, v1.REQ-007-AC-002
  _测试类型: E2E

---

- [ ] **TASK-012 · 前端静态资源 Embed 内嵌与单二进制集成验收**

  > 依赖：TASK-006, TASK-007, TASK-010, TASK-011

  - [ ] 配置 Go `embed.FS` 内嵌前端编译产物 `web/dist`
  - [ ] 编写 `cmd/tmux-ui/main.go` 启动主入口与 CLI 参数解析（`--port`, `--token`, `--socket-name`, `--socket-path`）
  - [ ] 构建独立单二进制可执行文件并进行全链路冒烟验收

  **验证方式：**
  ```bash
  pnpm --prefix web build && go build -o bin/tmux-ui ./cmd/tmux-ui && ./bin/tmux-ui --help
  ```

  **验收证据：** 成功编译产出无外部依赖的单二进制文件，验证 CLI 参数解析正常与本地服务一键启动。

  _需求: `requirements.md` 中 v1.REQ-001, v1.REQ-005, v1.REQ-006
  验收标准：v1.REQ-001-AC-005, v1.REQ-005-AC-001, v1.REQ-006-AC-001
  _测试类型: E2E

---

## 进度汇总

| TASK ID | 名称 | 测试类型 | 状态 | 关联需求 |
|---------|------|:-------:|:----:|---------|
| TASK-001 | 项目基础脚手架、配置模型与解析器 | Unit | 🟢 done | v1.REQ-004 |
| TASK-002 | Tmux CLI 采集器与版本探测兼容 | Unit | 🟢 done | v1.REQ-001 |
| TASK-003 | 跨平台进程树探测器与 Agent 进程关联 | Unit | 🟢 done | v1.REQ-001 |
| TASK-004 | 终端 ANSI 输出捕获与哈希差异跟踪 | Unit | 🟢 done | v1.REQ-002 |
| TASK-005 | 多智能体状态判定引擎与内存存储 | Unit | 🟢 done | v1.REQ-003 |
| TASK-006 | 后端 HTTP REST API 与 Token 认证中间件 | API | 🟢 done | v1.REQ-005 |
| TASK-007 | WebSocket Hub 实时推流服务 | API | ⬜ 待开始 | v1.REQ-005 |
| TASK-008 | 前端 Vite 脚手架与 WebSocket 通信 Hook | Unit | ⬜ 待开始 | v1.REQ-006 |
| TASK-009 | Web 总览卡片墙与多维状态筛选 | E2E | ⬜ 待开始 | v1.REQ-006 |
| TASK-010 | 终端详情页与 xterm.js ANSI 渲染器 | E2E | ⬜ 待开始 | v1.REQ-007 |
| TASK-011 | 状态变迁时间线抽屉与浏览器交互通知 | E2E | ⬜ 待开始 | v1.REQ-006, v1.REQ-007 |
| TASK-012 | 前端静态资源 Embed 内嵌与单二进制集成验收 | E2E | ⬜ 待开始 | v1.REQ-001, v1.REQ-005, v1.REQ-006 |

