# 需求文档（Requirements）

> 文件路径：`docs/spec/versions/v2/requirements.md`
> 适用版本：v2 · 生成日期：2026-11-12
> 状态：已定稿（Approved）
> 前置宪章：沿用 `docs/spec/versions/v1-finished/constitution.md` 与项目级 Steering 规则

---

## 1. 介绍

### 1.1 项目目标与背景
tmux-ui v1 版本成功构建了多智能体拓扑无侵入采集、ANSI 输出实时捕获、规则推断决策链、REST/WebSocket 推流、总览卡片墙与只读 xterm.js 终端查看器，并实现单二进制内嵌分发。

进入 **v2 版本**，重点推进两大核心能力扩展：
1. **历史输出持久化与时间轴回放（Output History & Timeline Playback）**：将智能体状态跃迁事件与终端输出快照关联存储，支持在 Web 端拖动时间轴或点击历史事件节点回溯当时的终端 ANSI 渲染状态，并支持一键恢复至实时流。
2. **多设备全站响应式设计与移动端体验优化（Mobile & Multi-Device Responsive Design）**：重塑 Web 控制台在小屏幕（手机、折叠屏、平板）上的布局体验，提供移动端横向滚动筛选栏、自适应单列卡片、触控友好的底部抽屉（Bottom Sheet），并优化 xterm.js 在狭窄视口下的自适应缩放与单手操作热区。

### 1.2 核心用户角色
- **移动/随时随地运维者 (Mobile Agent Operator)**：在离开工作台时，通过手机浏览器实时查看所有后台智能体的运行健康状况，排查停滞或异常。
- **智能体审计与回溯者 (Agent Auditor / Debugger)**：当智能体报错或行为偏离预期时，通过时间轴回放功能回溯复盘关键决策点的终端上下文。

---

## 2. 需求列表

---

### 需求 v2.REQ-001 · 终端历史输出持久化与快照索引引擎 (History Engine)

**用户故事：** 作为 Agent Auditor，我希望系统持久化记录每个 pane 的历史终端输出片段及状态变迁关联的快照，以便在 Agent 运行完成后或离线后仍能回溯过去的输出详情。

#### 验收标准

```yaml
- id: v2.REQ-001-AC-001
  ears: >
    When a pane experiences a state transition event or periodic checkpoint interval,
    the History Engine shall capture the current ANSI output buffer and associate it with a unique snapshot_id, timestamp, and event metadata.
  test_type: Unit
  expected:
    return_value: "SnapshotRecord with valid ID, Unix timestamp, PaneID, and non-empty ANSI payload"
    throws: ""
    side_effects:
      - "Snapshot indexed and queryable in local storage"

- id: v2.REQ-001-AC-002
  ears: >
    When total stored snapshots for a pane exceed the configured retention limit (default 200 snapshots or 10MB per pane),
    the History Engine shall prune the oldest snapshots in FIFO order without corrupting current timeline state.
  test_type: Unit
  expected:
    return_value: "int prunedCount >= 1"
    throws: ""
    side_effects:
      - "Storage size strictly bounded within threshold"

- id: v2.REQ-001-AC-003
  ears: >
    When the backend restarts or a pane is destroyed,
    the History Engine shall preserve history snapshots in local data store (e.g. JSONL append-log or lightweight SQLite) for the duration of the configured retention TTL.
  test_type: Unit
  expected:
    return_value: "[]SnapshotRecord correctly reloaded after store re-initialization"
    throws: ""
    side_effects: []
```

---

### 需求 v2.REQ-002 · 历史快照与回放 REST API (History REST API)

**用户故事：** 作为前端应用，我希望通过标准 HTTP API 查询指定 pane 的历史快照列表以及特定时刻的终端输出内容，以便驱动时间轴回放界面渲染。

#### 验收标准

```yaml
- id: v2.REQ-002-AC-001
  ears: >
    When a client sends a GET request to `/api/v1/panes/{id}/snapshots`,
    the Backend shall return a chronological list of historical snapshot metadata (snapshot_id, timestamp, event_type, old_state, new_state, summary).
  test_type: API
  expected:
    http_status: 200
    body_schema:
      code: "int, 0"
      message: "string, 'success'"
      data: "[]SnapshotMetadata, sorted by timestamp ascending"
    side_effects: []

- id: v2.REQ-002-AC-002
  ears: >
    When a client sends a GET request to `/api/v1/panes/{id}/snapshots/{snapshot_id}`,
    the Backend shall return the full raw ANSI terminal output associated with that snapshot.
  test_type: API
  expected:
    http_status: 200
    body_schema:
      code: "int, 0"
      message: "string, 'success'"
      data:
        snapshot_id: "string"
        timestamp: "int64"
        pane_id: "string"
        output: "string with preserved ANSI codes"
    side_effects: []

- id: v2.REQ-002-AC-003
  ears: >
    When a client requests a snapshot with an invalid pane ID or non-existent snapshot ID,
    the Backend shall return HTTP 404 with a standard JSON error envelope.
  test_type: API
  expected:
    http_status: 404
    body_schema:
      error:
        code: "string, 'SNAPSHOT_NOT_FOUND' or 'PANE_NOT_FOUND'"
        message: "string"
    side_effects: []
```

---

### 需求 v2.REQ-003 · 终端详情页历史回溯与时间轴播放器 (Playback UI & Live Scrubber)

**用户故事：** 作为 Agent Operator，我希望在终端详情页可以通过拖动时间轴滑块或点击时间线事件节点，瞬间回溯查看当时的终端输出，并能一键返回实时最新输出（Live）。

#### 验收标准

```yaml
- id: v2.REQ-003-AC-001
  ears: >
    When navigating to the terminal detail view,
    the UI shall render an interactive Playback Timeline bar displaying historical state transition markers alongside the live terminal.
  test_type: E2E
  expected:
    ui_state: "Playback bar visible with draggable slider and clickable event milestones"
    url: "/panes/:id"
    side_effects: []

- id: v2.REQ-003-AC-002
  ears: >
    When the user drags the timeline slider or clicks a historical milestone,
    the UI shall render the historical ANSI output in xterm.js, display a prominent 'Historical Playback' banner, and highlight the selected timestamp.
  test_type: E2E
  expected:
    ui_state: "xterm.js displaying snapshot output with 'Historical Playback' indicator banner and active milestone highlight"
    url: "/panes/:id"
    side_effects: []

- id: v2.REQ-003-AC-003
  ears: >
    While in historical playback mode,
    when the user clicks the 'Return to Live' button,
    the UI shall seamlessly restore real-time WebSocket output streaming and dismiss the historical banner.
  test_type: E2E
  expected:
    ui_state: "Live status restored, historical banner dismissed, latest real-time ANSI stream visible"
    url: "/panes/:id"
    side_effects: []
```

---

### 需求 v2.REQ-004 · 全站移动端与多设备响应式自适应设计 (Mobile Responsive Layout)

**用户故事：** 作为使用手机或平板的 Agent Operator，我希望页面能够自适应小屏幕视口，导航栏与过滤栏紧凑且易于单手操作，以便随时随地使用移动设备巡检各个 Agent 的状态。

#### 验收标准

```yaml
- id: v2.REQ-004-AC-001
  ears: >
    When viewport width is less than 768px (e.g. mobile 375px~430px),
    the Header component shall collapse secondary elements into a compact mobile layout with zero horizontal overflow.
  test_type: E2E
  expected:
    ui_state: "Header fits viewport width without horizontal scrollbar, status badge and brand title cleanly aligned"
    url: "/"
    side_effects: []

- id: v2.REQ-004-AC-002
  ears: >
    When viewed on a mobile device (viewport < 768px),
    the FilterBar shall convert to a smooth horizontally scrollable tab list with touch momentum, and the search input shall expand to full container width.
  test_type: E2E
  expected:
    ui_state: "Filter tabs horizontally scrollable with visible active counters; search box full width with touch target >= 44px"
    url: "/"
    side_effects: []

- id: v2.REQ-004-AC-003
  ears: >
    When viewed on a mobile device (viewport < 768px),
    the OverviewGrid shall arrange pane cards in a single-column stacked layout with responsive padding and readable typography.
  test_type: E2E
  expected:
    ui_state: "Cards rendered as 1 column, status indicators, badges, and ANSI preview wrap gracefully without clipping"
    url: "/"
    side_effects: []
```

---

### 需求 v2.REQ-005 · 移动端抽屉与终端触控交互适配 (Mobile Touch & Drawer Adaptation)

**用户故事：** 作为移动端用户，我希望在手机上查看时间线抽屉和终端详情时，交互形式符合移动端交互习惯，按钮具有充足的触控热区。

#### 验收标准

```yaml
- id: v2.REQ-005-AC-001
  ears: >
    While on a mobile viewport (width < 768px),
    when the user opens the Timeline Drawer,
    the component shall slide up from the bottom as a Bottom Sheet with maximum height 85vh, supporting backdrop tap or swipe-to-dismiss.
  test_type: E2E
  expected:
    ui_state: "Bottom Sheet slides up smoothly with drag handle, backdrop blur, and accessible touch close button"
    url: "/"
    side_effects: []

- id: v2.REQ-005-AC-002
  ears: >
    While viewing terminal detail on a mobile screen,
    the TerminalViewer shall dynamically re-fit font size and dimensions via FitAddon, provide a floating 'Scroll to Bottom' button, and ensure all interactive buttons meet >= 44x44px touch targets.
  test_type: E2E
  expected:
    ui_state: "xterm canvas perfectly bounds within viewport, floating quick-action buttons accessible for one-thumb use"
    url: "/panes/:id"
    side_effects: []
```

---

## 3. 质量检查清单

- [x] 每条 EARS 中“系统响应”具备客观可验证性（无主观形容词）。
- [x] 覆盖了异常场景（404 找不到快照、非法 ID、存储溢出 FIFO 淘汰）。
- [x] 每条 AC 均明确关联了 `test_type`（Unit / API / E2E）。
- [x] 需求编号与 AC 编号连续规范，统一使用 `v2.REQ-XXX` 前缀隔离 v1 历史需求。
- [x] 包含 Playwright 视觉回归与移动端视口（iPhone/Android 宽度）明确测试要求。
