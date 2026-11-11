# 实施计划（Tasks）

> 文件路径：`docs/spec/versions/v2/tasks.md`
> 适用版本：v2 · 生成日期：2026-11-12
> 前置依据：`docs/spec/versions/v2/requirements.md` 与 `docs/spec/versions/v2/design.md`

执行时须严格遵循需求文档与验收标准，严格遵循测试驱动开发（TDD）红-绿-重构循环。
每个任务必须执行真实命令验证并保留测试通过证据。

---

## 任务清单

---

- [x] **TASK-013 · 历史快照数据模型与本地 JSONL 持久化引擎**

  > 依赖：无

  - [x] 在 `internal/model` 中定义 `SnapshotMetadata` 与 `SnapshotRecord` 领域实体
  - [x] 在 `internal/config` 中扩展加载 `history` 配置节点，提供开箱即用默认值降级
  - [x] 在 `internal/history` 中实现 `Storage` 模块：支持路径安全清洗（转义 `%` 等字符，防止路径穿越）
  - [x] 实现结构化追加写入 `.jsonl`、多协程并发写入锁安全与单 Pane 读写分离
  - [x] 实现 FIFO 滚动淘汰与自动裁剪机制（当记录数超过 `max_snapshots_per_pane` 或文件大小超限时原子重写文件）

  **验证方式：**
  ```bash
  go test -v ./internal/history/... -run TestStorage
  ```

  **验收证据：** 单元测试 100% 通过，成功验证单行追加写入、并发安全锁、特殊字符路径清洗与超限自动 FIFO 截断。

  _需求: `requirements.md` 中 v2.REQ-001
  验收标准：v2.REQ-001-AC-001, v2.REQ-001-AC-002, v2.REQ-001-AC-003
  _测试类型: Unit

---

- [x] **TASK-014 · 状态变迁快照挂钩与周期性保底 Checkpoint**

  > 依赖：TASK-013

  - [x] 在 `internal/history` 中实现 `HistoryEngine`：挂载至 `StateStore` 变迁监听通道
  - [x] 当目标 Pane 发生状态变迁事件（`state_changed`）时，自动触发创建快照并持久化
  - [x] 实现周期性定时器（Checkpoint Ticker，默认 30s）：输出内容有哈希变动但未触发状态跃迁时生成保底快照
  - [x] 实现 Pane 销毁后的快照缓存安全保留与保留期 TTL 检查

  **验证方式：**
  ```bash
  go test -v ./internal/history/... -run TestEngine
  ```

  **验收证据：** 单元测试通过，模拟状态变迁与定时 Checkpoint 触发成功生成有序快照流，验证销毁 Pane 不丢失已持久化记录。

  _需求: `requirements.md` 中 v2.REQ-001
  验收标准：v2.REQ-001-AC-001, v2.REQ-001-AC-003
  _测试类型: Unit

---

- [ ] **TASK-015 · 历史快照 REST API 与 WebSocket 广播扩展**

  > 依赖：TASK-014

  - [ ] 在 `internal/server/http_server.go` 中注册路由 `GET /api/v1/panes/{id}/snapshots`（支持 `since`、`limit` 参数及宽容路由参数解析）
  - [ ] 注册路由 `GET /api/v1/panes/{id}/snapshots/{snapshot_id}`（获取特定历史时刻完整 ANSI 输出与元数据）
  - [ ] 实现无效 Pane ID 或不存在 Snapshot ID 的标准 404 JSON 错误信封响应
  - [ ] 扩展 `ws_hub.go` 在快照生成时向订阅客户端广播 `snapshot_created` 事件

  **验证方式：**
  ```bash
  go test -v ./internal/server/... -run TestSnapshotAPI
  ```

  **验收证据：** API 测试展示 200 / 404 真实输出，验证快照元数据列表查询、单条 ANSI 内容读取及 WebSocket 广播帧下发正常。

  _需求: `requirements.md` 中 v2.REQ-002
  验收标准：v2.REQ-002-AC-001, v2.REQ-002-AC-002, v2.REQ-002-AC-003
  _测试类型: API

---

- [ ] **TASK-016 · 移动端视口探测 Hook 与响应式卡片墙重塑**

  > 依赖：TASK-015

  - [ ] 实现 `web/src/hooks/useMediaQuery.ts`，提供可复用的小屏视口侦测能力（`< 768px` 判定）
  - [ ] 重构 `web/src/App.tsx` 顶部 Header：在移动端紧凑排布标题与状态指示器，消除任何横向滚动溢出
  - [ ] 重构 `web/src/components/FilterBar.tsx`：支持触控横向平滑滚动标签（Touch Scroll），搜索输入框在手机端自适应 100% 宽度
  - [ ] 重构 `web/src/components/OverviewGrid.tsx` 与 `PaneCard.tsx`：小屏幕下自动平滑过渡为单列紧凑卡片，标签与 ANSI 预览换行对齐

  **验证方式：**
  ```bash
  pnpm --prefix web test --run && playwright test mobile_grid
  ```

  **验收证据：** 组件测试通过，Playwright 截图留存桌面端（1280px）与移动端（390px iPhone）卡片墙视觉渲染证据。

  _需求: `requirements.md` 中 v2.REQ-004
  验收标准：v2.REQ-004-AC-001, v2.REQ-004-AC-002, v2.REQ-004-AC-003
  _测试类型: E2E

---

- [ ] **TASK-017 · 移动端 Bottom Sheet 抽屉与终端触控交互优化**

  > 依赖：TASK-016

  - [ ] 开发轻量 `web/src/components/BottomSheet.tsx`：自底向上滑出动画、最大高度 85vh 约束、顶部拖拽手柄、点击遮罩或下滑关闭
  - [ ] 重构 `web/src/components/TimelineDrawer.tsx`：移动端视口下无缝转换为 Bottom Sheet 呈现时间线节点
  - [ ] 优化 `web/src/components/TerminalViewer.tsx`：窄屏视口下动态调整 xterm 容器与 FitAddon，添加悬浮「滚动至底部」快捷按钮
  - [ ] 确保移动端所有可点击元素满足 $\ge 44 \times 44\text{px}$ 触控热区可访问性规范

  **验证方式：**
  ```bash
  pnpm --prefix web test --run && playwright test mobile_bottomsheet
  ```

  **验收证据：** 组件单元测试通过，Playwright 截图展示移动端 Bottom Sheet 弹出与终端浮动操作按钮交互视图。

  _需求: `requirements.md` 中 v2.REQ-005
  验收标准：v2.REQ-005-AC-001, v2.REQ-005-AC-002
  _测试类型: E2E

---

- [ ] **TASK-018 · 终端详情页时间轴回放播放器与 Live 追踪切换**

  > 依赖：TASK-015, TASK-017

  - [ ] 实现 `web/src/hooks/useHistory.ts`：拉取指定 Pane 的历史快照列表，维护回放模式与选中快照
  - [ ] 开发 `web/src/components/PlaybackBar.tsx`：可拖动时间轴滑块、历史状态跃迁节点刻度渲染、快照时间指示
  - [ ] 终端详情页联动：拖拽或点击刻度时暂停实时 WS 推流，xterm.js 瞬时切换渲染历史 ANSI 文本，并展示「Historical Playback」提示条
  - [ ] 点击「Return to Live」按钮一键退出回放模式，无缝恢复实时终端流更新

  **验证方式：**
  ```bash
  pnpm --prefix web test --run && playwright test playback
  ```

  **验收证据：** 前端单元测试通过，Playwright 截图展示时间轴拖动回溯历史终端输出与一键恢复 Live 状态视图。

  _需求: `requirements.md` 中 v2.REQ-003
  验收标准：v2.REQ-003-AC-001, v2.REQ-003-AC-002, v2.REQ-003-AC-003
  _测试类型: E2E

---

- [ ] **TASK-019 · 前端静态产物内嵌与 v2 全链路回归验收**

  > 依赖：TASK-013, TASK-014, TASK-015, TASK-016, TASK-017, TASK-018

  - [ ] 执行前端生产打包 `pnpm --prefix web build`
  - [ ] 构建包含 v2 新特性的 Go 单二进制程序 `bin/tmux-ui`
  - [ ] 执行端到端全链路冒烟测试与桌面/移动端双视口视觉回归测试
  - [ ] 汇总生成 `docs/spec/versions/v2/reports/acceptance-report-{{DATE}}.md` 交付报告

  **验证方式：**
  ```bash
  pnpm --prefix web build && go build -o bin/tmux-ui ./cmd/tmux-ui && playwright test
  ```

  **验收证据：** 成功构建单二进制可执行文件，全链路回归测试 100% 通过，生成完整质量验收交付报告。

  _需求: `requirements.md` 中全部 v2.REQ-001 ~ v2.REQ-005
  验收标准：全部 v2 AC 覆盖
  _测试类型: E2E

---

## 进度汇总

| TASK ID | 名称 | 测试类型 | 状态 | 关联需求 |
|---------|------|:-------:|:----:|---------|
| TASK-013 | 历史快照数据模型与本地 JSONL 持久化引擎 | Unit | 🟢 done | v2.REQ-001 |
| TASK-014 | 状态变迁快照挂钩与周期性保底 Checkpoint | Unit | 🟢 done | v2.REQ-001 |
| TASK-015 | 历史快照 REST API 与 WebSocket 广播扩展 | API | ⚪ 待开始 | v2.REQ-002 |
| TASK-016 | 移动端视口探测 Hook 与响应式卡片墙重塑 | E2E | ⚪ 待开始 | v2.REQ-004 |
| TASK-017 | 移动端 Bottom Sheet 抽屉与终端触控交互优化 | E2E | ⚪ 待开始 | v2.REQ-005 |
| TASK-018 | 终端详情页时间轴回放播放器与 Live 追踪切换 | E2E | ⚪ 待开始 | v2.REQ-003 |
| TASK-019 | 前端静态产物内嵌与 v2 全链路回归验收 | E2E | ⚪ 待开始 | v2.REQ-001 ~ v2.REQ-005 |
