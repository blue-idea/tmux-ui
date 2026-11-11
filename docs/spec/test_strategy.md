# 测试策略简报（Test Strategy）

> 文件路径：`docs/spec/test_strategy.md`
> 参考方法论：`phases/qa_engine.md` §第1阶段
> 创建步骤：STEP 5（拆分任务，首次进入时生成）
> 生成日期：2026-11-11

---

## 1. 项目信息

```yaml
project:
  name: "tmux-ui"
  type: web-app
  languages: [go, typescript]
  frameworks: [react, chi, xterm.js]
```

---

## 2. 风险概况

```yaml
risk_profile:
  data_sensitivity: low        # 本地开发辅助工具，仅处理终端回显与进程元数据
  user_impact: internal        # 面向开发者与智能体运维人员
  deployment_frequency: weekly
  regulatory: [none]
```

---

## 3. 测试范围

```yaml
test_scope:
  in_scope:
    - "tmux CLI 命令构造与格式解析（含 -L/-S socket 选项与 tmux -V 自适应）"
    - "跨平台 ps 命令子进程树探测与关联"
    - "终端 ANSI 输出捕获、64位哈希计算与活跃度追踪"
    - "多智能体规则研判引擎（优先级判定链与正则匹配）"
    - "线程安全内存状态存储（StateStore）与环形缓冲时间线（RingBuffer）"
    - "HTTP REST 路由（/api/v1/health, /api/v1/panes, /api/v1/panes/{id} 宽容路由）与 Token 鉴权"
    - "WebSocket Hub 握手、全量快照与增量广播推流"
    - "React 前端总览卡片墙、状态灯视觉指示与多维状态筛选"
    - "xterm.js ANSI 彩色终端挂载渲染、FitAddon 尺寸自适应与只读保护"
    - "状态跃迁至 waiting_input / error 时的桌面/浏览器通知"

  out_of_scope:
    - "从 Web 端向 tmux pane 发送输入（send-keys）：MVP 阶段明确排除写操作以保证执行安全"
    - "跨机器多节点 SSH 隧道汇聚监控：属于后续演进版本规划范围"
```

---

## 4. 测试类型决策

| 测试类型 | 是否执行 | 覆盖目标 | 工具 / 框架 |
|---------|:-------:|---------|------------|
| **单元测试** | ✅ 执行 | 后端核心逻辑行覆盖率 ≥ 80%，状态机判定 100%；前端 Hooks 覆盖率 ≥ 80% | Go `testing` + `testify` / `vitest` |
| **集成 / API 测试** | ✅ 执行 | 所有 REST 接口（正常/鉴权拦截/404）及 WebSocket 握手推流 | Go `httptest` + `gorilla/websocket` |
| **E2E / 视觉回归** | ✅ 执行 | 覆盖 4 条关键用户旅程（总览墙浏览、状态筛选、终端详情渲染、通知下发），**必须截图留证** | **Playwright MCP** |
| **性能基准测试** | ✅ 执行 | 终端正则匹配与哈希计算单次耗时 < 1ms | Go `testing.B`（`go test -bench`） |
| **安全性测试** | ✅ 执行 | 本地回环绑定校验、命令注入防范、Token 鉴权阻断 | 自动化集成用例 |

---

## 5. 测试金字塔目标比例

```
      /  E2E  \          10% — 关键用户旅程与视觉回归截图
     / 集成测试 \         25% — API 契约、WebSocket 推流、并发存储
    /   单元测试   \       65% — tmux 输出解析、状态机判定、哈希 Diff、React 基础组件
```

---

## 6. 环境配置

```yaml
environments:
  dev:
    url: "http://127.0.0.1:8080"
    db: "in-memory"
  staging:
    url: "http://127.0.0.1:8089"
    db: "in-memory (mock tmux data)"
  prod:
    url: "http://127.0.0.1:8080"
    smoke_only: true
```

---

## 7. 测试数据策略

| 测试层级 | 数据来源 | 清理策略 |
|---------|---------|---------|
| 后端单元测试 | Mock Tmux Executor / 纯字符串输入 | 无需清理（纯内存/模拟环境） |
| 集成 / API 测试 | 预置测试 Session（从 `docs/spec/info.md` 提取） | 测试套件完成后重置内存 Store |
| E2E 测试 | Playwright 驱动测试页面并连接 Mock 后端 | 每个用例执行后自动清理上下文 |

> ⚠️ 核心规则：测试账号、Token 和测试 Tmux 会话数据统一从 `docs/spec/info.md` 获取，严禁随意捏造。

---

## 8. 质量基线（最低通过标准）

| 指标 | 目标值 | 验证方式 |
|------|-------|---------|
| 后端行覆盖率 | ≥ 80% | `go test -coverprofile=coverage.out ./...` |
| 状态机与匹配覆盖率 | 100% | 覆盖所有状态跃迁分支与 Agent 正则 |
| 前端类型检查错误数 | 0 | `pnpm exec tsc --noEmit` |
| 前端单元测试通过率 | 100% | `pnpm test --run` |
| UI 视觉回归截图留证 | 100% 存在 | Playwright 产出 Baseline 与 Diff 截图 |
| 缺陷逃逸率 | 0（无阻塞 P0/P1 缺陷上线） | 综合验收审查 |

