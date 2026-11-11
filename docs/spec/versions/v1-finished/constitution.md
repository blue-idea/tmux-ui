# 项目宪法（Constitution）

> 文件路径：`docs/spec/versions/v1/constitution.md`
> 适用版本：v1 · 生成日期：2026-11-11

---

## 1. 项目说明

**tmux 多智能体可视化工具 (tmux-ui)** 是一个轻量级、无侵入式的多智能体运行状态可视化系统。
系统通过调用本机 `tmux` 命令行接口（`list-panes`, `capture-pane` 等）采集终端拓扑与实时输出，利用规则引擎与启发式算法研判多智能体状态（`working`, `idle`, `waiting_input`, `error`, `exited`），并通过内置 Web 服务向前端实时推流，提供总览卡片墙与基于 `xterm.js` 的终端交互详情界面。

---

## 2. 技术栈约束

- **后端主语言**：Go (`go 1.22+`)，单一静态二进制分发，使用 Go 标准库 `embed` 打包编译后的前端静态资源。
- **前端框架**：React 18/19 + Vite + TypeScript + Tailwind CSS。
- **终端渲染**：`@xterm/xterm`，完整支持 ANSI 彩色与终端控制序列渲染。
- **配置格式**：YAML 格式，存储于 `config/agents.yaml`，统一集中管理各类 Agent 的识别特征与状态正则规则。
- **通信协议**：HTTP REST API（获取静态拓扑与配置） + WebSocket（实时双向推流与状态变化通知）。

### TypeScript 规范

- 必须启用严格模式（`strict: true`）。
- 严禁使用 `any`、`unknown` 及 `as` 强制类型断言；禁止包含忽略编译警告的注释（如 `@ts-ignore`）。
- 前端任务提交前必须自动运行 `pnpm exec tsc --noEmit` 校验。

---

## 3. 包管理规范

1. 前端统一使用 **pnpm**，严格锁定依赖版本（`pnpm-lock.yaml` 必须提交），严禁混用 `npm` 或 `yarn`。
2. 后端统一使用 **Go Modules**（`go.mod` 与 `go.sum` 必须提交）。
3. 生产依赖必须使用稳定版本，禁止引入任何 `alpha` / `beta` / `rc` 等预览版依赖包。

---

## 4. 编码安全与权限边界

### 网络监听与隔离

- 后端服务默认仅绑定监听本地回环地址 `127.0.0.1`，严禁在未配置安全鉴权的情况下公开暴露至 `0.0.0.0`。
- 支持启动时生成或指定随机安全 Token，用于 Web 访问和 WebSocket 握手鉴权。

### 命令执行安全（只读约束）

- MVP 阶段系统严格坚持**只读监控**原则，只调用 `tmux list-panes`、`tmux capture-pane` 及安全的兼容进程树检测 `ps -axo`。
- 严禁拼接任意未经转义的用户输入作为 shell 命令参数。后续若引入 `send-keys` 写操作，必须设置显式配置白名单并经双重安全开关启用。

### 敏感数据防护

- 敏感参数或自定义端口必须通过 `.env` 或环境变量读取，禁止在代码库中硬编码。
- 终端捕获内容若涉及敏感输出应具备脱敏过滤能力，且 `.env` 必须写入 `.gitignore`。

---

## 5. 质量基线与开发方式

### 严格遵循测试驱动开发（TDD）

```
⛔ 没有失败的测试，禁止编写任何生产代码
```

1. **红（Red）**：编写针对目标特性的单元/集成测试，先运行验证其失败并确定失败原因符合预期。
2. **绿（Green）**：编写最小实现的生产代码，使测试通过。
3. **重构（Refactor）**：在测试用例的绿灯防护下，优化代码结构、消除重复与坏味道，确保代码整洁。

### 质量门槛

| 检查维度 | 工具 / 命令 | 验收标准 |
|---------|-----------|---------|
| Go 单元测试 | `go test -v -race ./...` | 100% 通过，无竞态告警，核心模块覆盖率 >= 80% |
| 前端静态类型检查 | `pnpm exec tsc --noEmit` | 0 Error |
| 前端单元测试 | `pnpm test --run` (Vitest) | 100% 通过 |
| 视觉回归与 E2E | Playwright MCP | UI 变更须截屏留存 Baseline 与 Diff，生成真实验收证据 |
| 跨平台换行符 | `.gitattributes` | 强制文本文件 LF (`eol=lf`)，消除脏提交 |

---

## 6. 文档结构规划

```
docs/
├── AGENTS.md                       # 项目级 Agent 入口
└── spec/
    ├── README.md                   # 版本索引（current_version: v1）
    ├── test_strategy.md            # 测试策略简报（项目级）
    ├── info.md                     # 测试账号与环境信息（项目级，gitignore）
    ├── steering/                   # 项目级恒定规则（不随 version 变更）
    │   ├── project-rules.md        # 协作 / Git / 测试门槛
    │   ├── tech-stack.md           # 锁定技术栈
    │   └── style-guide.md          # 命名 / 错误格式 / 日志
    └── versions/
        ├── v1/
        │   ├── constitution.md     # 本文件（v1 项目宪法）
        │   ├── requirements.md     # 需求与 EARS 验收标准 (STEP 2)
        │   ├── design.md           # 技术设计概览与架构 (STEP 3)
        │   ├── data.md             # 状态模型与规则定义 (STEP 3 按需)
        │   ├── api.md              # REST / WebSocket 协议规范 (STEP 3 按需)
        │   ├── tasks.md            # 任务拆分清单 (STEP 5)
        │   ├── traceability.md     # 可追溯性矩阵 (STEP 5)
        │   ├── analysis-report-{{DATE}}.md # 一致性审查报告 (STEP 4)
        │   ├── ac/                 # 各任务 AC 验收矩阵
        │   │   └── TASK-XXX-AC.md
        │   └── reports/            # 验收与测试报告
        │       └── REPORT_ID-report.md
        └── v2/                     # v1 完成后新建
            └── ...
```

