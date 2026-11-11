# 需求文档（Requirements）

> 文件路径：`docs/spec/versions/v1/requirements.md`
> 适用版本：v1 · 生成日期：2026-11-11
> 状态：已定稿（Approved）

---

## 1. 介绍

### 1.1 项目目标
**tmux-ui** 是一个针对本地/远端多智能体（如 Claude Code, Codex, Gemini CLI, Aider 等）并发运行场景的可视化监控工具。通过无侵入式地调用 `tmux` CLI 和进程探测命令，实时采集终端拓扑与输出，利用可配置规则引擎推断每个 Agent 的运行状态（`working`, `idle`, `waiting_input`, `error`, `exited`），并通过轻量级 Web 服务提供总览卡片墙与基于 `xterm.js` 的终端交互详情界面。

### 1.2 核心用户角色
- **智能体开发者 / 运维者 (Agent Operator)**：同时运行多个自主编码或对话智能体，需要宏观掌控全局执行进度、识别停滞或等待人工介入的智能体，并快速下钻排查异常。

---

## 2. 需求列表

---

### 需求 v1.REQ-001 · 拓扑与进程树无侵入采集 (Collector)

**用户故事：** 作为 Agent Operator，我希望系统自动发现当前机器上所有 tmux session、window 和 pane，并获取 pane 内当前进程信息，以便全景掌握所有并发运行的智能体实例。

#### 验收标准

```yaml
- id: v1.REQ-001-AC-001
  ears: >
    When the collector polling ticker triggers (every 500ms~1000ms),
    the Collector shall execute `tmux list-panes -a -F ...` and parse session_name, window_id, window_name, pane_id, pane_pid, pane_current_command, and pane_dead.
  test_type: Unit
  expected:
    return_value: "[]PaneInfo, len >= 0, fields correctly parsed from tmux format output"
    throws: ""
    side_effects: []

- id: v1.REQ-001-AC-002
  ears: >
    When tmux pane PIDs are collected,
    the Collector shall query process command trees using cross-platform compatible syntax `ps -axo pid=,ppid=,command=` and associate child processes with each pane.
  test_type: Unit
  expected:
    return_value: "map[string]ProcessInfo, accurately identifying process hierarchy under pane PID"
    throws: ""
    side_effects: []

- id: v1.REQ-001-AC-003
  ears: >
    If the tmux server is not running or socket is unreachable,
    then the Collector shall catch the execution error and return a typed TMUX_UNAVAILABLE error without process crash.
  test_type: Unit
  expected:
    return_value: "nil"
    throws: "ErrTmuxUnavailable"
    side_effects: []

- id: v1.REQ-001-AC-004
  ears: >
    When the backend initializes,
    the Collector shall query `tmux -V` to detect version features and gracefully downgrade format variables for compatibility.
  test_type: Unit
  expected:
    return_value: "TmuxVersionInfo, Major >= 2"
    throws: ""
    side_effects: []

- id: v1.REQ-001-AC-005
  ears: >
    When custom socket options are provided via CLI flags (`--socket-name` / `--socket-path`) or configuration file,
    the Collector shall append `-L <name>` or `-S <path>` to all tmux commands, with CLI flags taking precedence over configuration.
  test_type: Unit
  expected:
    return_value: "TmuxCommandArgs containing specified socket flags"
    throws: ""
    side_effects: []
```

---

### 需求 v1.REQ-002 · 终端输出捕获与哈希差异推断 (Capture & Stream)

**用户故事：** 作为 Agent Operator，我希望系统能够持续捕获目标 pane 的终端输出（保留 ANSI 颜色）并维护内容变更差异，以便展示实时终端预览与推断活跃状态。

#### 验收标准

```yaml
- id: v1.REQ-002-AC-001
  ears: >
    When capturing output for a specified pane ID,
    the Collector shall invoke `tmux capture-pane -p -e -t %id` with ANSI escape sequence retention.
  test_type: Unit
  expected:
    return_value: "string containing raw terminal bytes with ANSI escape codes"
    throws: ""
    side_effects: []

- id: v1.REQ-002-AC-002
  ears: >
    When new pane output is captured,
    the StateStore shall compute a 64-bit content hash of the trailing output and record the last activity timestamp if hash changes.
  test_type: Unit
  expected:
    return_value: "bool (hasChanged), updated LastActiveAt timestamp"
    throws: ""
    side_effects:
      - "In-memory StateStore updates pane output cache"

- id: v1.REQ-002-AC-003
  ears: >
    If a pane is dead or exited,
    the Capture module shall mark the pane buffer as finalized and cease further polling for this pane.
  test_type: Unit
  expected:
    return_value: "PaneState == exited"
    throws: ""
    side_effects: []
```

---

### 需求 v1.REQ-003 · 多智能体状态判定引擎 (Detector & State Engine)

**用户故事：** 作为 Agent Operator，我希望系统依据输出特征正则、活跃度及进程状态自动推断智能体状态（`working`, `idle`, `waiting_input`, `error`, `exited`），以便及时知晓智能体何时需要我交互或已完成退出。

#### 验收标准

```yaml
- id: v1.REQ-003-AC-001
  ears: >
    When a pane's `pane_dead` flag is equal to 1,
    the Detector shall evaluate the pane state as `exited` with highest priority.
  test_type: Unit
  expected:
    return_value: "StateExited"
    throws: ""
    side_effects: []

- id: v1.REQ-003-AC-002
  ears: >
    While the pane is alive,
    when the trailing N lines match configured agent rule regex (e.g. `Do you want to proceed?` or `(y/n)` for Claude Code),
    the Detector shall evaluate the pane state as `waiting_input`.
  test_type: Unit
  expected:
    return_value: "StateWaitingInput"
    throws: ""
    side_effects: []

- id: v1.REQ-003-AC-003
  ears: >
    While the pane is alive and trailing lines match working regex (e.g. `esc to interrupt`),
    or when output hash continues to change within N seconds,
    the Detector shall evaluate the pane state as `working`.
  test_type: Unit
  expected:
    return_value: "StateWorking"
    throws: ""
    side_effects: []

- id: v1.REQ-003-AC-004
  ears: >
    While the pane process is a generic shell (bash/zsh) with no running agent child processes,
    the Detector shall evaluate the pane state as `idle` or `exited`.
  test_type: Unit
  expected:
    return_value: "StateIdle"
    throws: ""
    side_effects: []

- id: v1.REQ-003-AC-005
  ears: >
    When a pane state transition occurs,
    the StateStore shall append a transition record (from_state, to_state, timestamp, matched_rule) to the pane's timeline history.
  test_type: Unit
  expected:
    return_value: "StateTimelineEvent appended, buffer bounded to max 100 entries per pane"
    throws: ""
    side_effects:
      - "Timeline event recorded in memory"
```

---

### 需求 v1.REQ-004 · 规则引擎配置与自定义扩展 (Configuration)

**用户故事：** 作为 Agent Operator，我希望通过 YAML 配置文件自定义不同 Agent（Claude Code, Codex, Aider 等）的识别规则，以便轻松适配新的智能体命令行工具。

#### 验收标准

```yaml
- id: v1.REQ-004-AC-001
  ears: >
    When the backend loads `config/agents.yaml`,
    the ConfigManager shall parse agent match patterns, process names, and state regular expressions.
  test_type: Unit
  expected:
    return_value: "AgentConfigList valid, containing default rules for claude-code, codex, gemini-cli, aider"
    throws: ""
    side_effects: []

- id: v1.REQ-004-AC-002
  ears: >
    If the configuration file is missing or invalid YAML,
    the ConfigManager shall fall back to built-in default agent rules and emit a structured warning log.
  test_type: Unit
  expected:
    return_value: "BuiltinFallbackConfig"
    throws: ""
    side_effects:
      - "Structured WARN log emitted"
```

---

### 需求 v1.REQ-005 · 后端 HTTP REST 与 WebSocket 推送服务 (Backend API & WS)

**用户故事：** 作为 Web 前端，我希望通过 HTTP 接口拉取全量拓扑并基于 WebSocket 接收增量变更与状态跃迁广播，以便实现实时无刷新的可视化呈现。

#### 验收标准

```yaml
- id: v1.REQ-005-AC-001
  ears: >
    When a client sends a GET request to `/api/v1/panes`,
    the HTTP Server shall respond with a JSON array of all active panes and their current states.
  test_type: API
  expected:
    http_status: 200
    body_schema:
      code: "number, equals 0"
      message: "string, equals 'success'"
      data: "array of PaneSummary objects containing pane_id, session_name, window_name, agent_type, state, idle_seconds"
    side_effects: []

- id: v1.REQ-005-AC-002
  ears: >
    When a client sends a GET request to `/api/v1/panes/:id` (supporting both bare numbers normalized to `%<number>` and URL-encoded `%25<number>`),
    the HTTP Server shall normalize the pane identifier and respond with detailed pane metadata, trailing output buffer, and state history timeline.
  test_type: API
  expected:
    http_status: 200
    body_schema:
      code: "number, equals 0"
      data: "PaneDetail object containing pane_id, ansi_content, state, timeline"
    side_effects: []

- id: v1.REQ-005-AC-003
  ears: >
    When a WebSocket client connects to `/ws`,
    the Server shall upgrade connection, send the initial snapshot, and stream subsequent topology/state change events.
  test_type: API
  expected:
    http_status: 101
    body_schema:
      event_type: "string, enum: [snapshot, pane_updated, state_changed, output_diff]"
      payload: "object"
    side_effects:
      - "Client registered in hub"

- id: v1.REQ-005-AC-004
  ears: >
    While auth token is configured,
    if a request arrives without valid `Authorization: Bearer <token>` or token query param,
    then the Server shall reject the request with HTTP 401.
  test_type: API
  expected:
    http_status: 401
    body_schema:
      error:
        code: "UNAUTHORIZED"
        message: "Authentication token is missing or invalid"
    side_effects: []
```

---

### 需求 v1.REQ-006 · Web 总览墙与状态筛选通知 (Frontend Overview)

**用户故事：** 作为 Agent Operator，我希望在 Web 浏览器上通过卡片墙一览所有智能体运行状态，按状态分类筛选，并在有 Agent 等待输入时获得浏览器桌面通知。

#### 验收标准

```yaml
- id: v1.REQ-006-AC-001
  ears: >
    When the Overview page renders,
    the UI shall display a responsive card grid where each card presents session/window name, pane ID, agent badge, status indicator lamp, trailing terminal snippet, and idle duration.
  test_type: E2E
  expected:
    ui_state: "Grid of pane cards rendered with matching status colors (green: working, yellow: waiting_input, gray: idle, red: error, dark: exited)"
    url: "/"
    side_effects: []

- id: v1.REQ-006-AC-002
  ears: >
    When the user selects a status filter tab (e.g. `waiting_input`),
    the UI shall filter the displayed cards to show only panes matching the selected state.
  test_type: E2E
  expected:
    ui_state: "Only cards with matching status are visible; card count badge updates accurately"
    url: "/?status=waiting_input"
    side_effects: []

- id: v1.REQ-006-AC-003
  ears: >
    When a pane's state transitions to `waiting_input` or `error` and notifications are granted,
    the Web Application shall trigger a browser notification with pane identifier and prompt snippet.
  test_type: E2E
  expected:
    ui_state: "Browser notification triggered with title '[tmux-ui] Agent Waiting for Input' and prompt body"
    side_effects: []
```

---

### 需求 v1.REQ-007 · 终端详情页与状态时间线 (Frontend Detail View & xterm.js)

**用户故事：** 作为 Agent Operator，我希望点击卡片进入详情页，通过 xterm.js 查看完整 ANSI 彩色终端输出并审视其状态跃迁历史时间线，以便精确排查交互上下文。

#### 验收标准

```yaml
- id: v1.REQ-007-AC-001
  ears: >
    When the user clicks on a pane card,
    the Application shall navigate to the detail view and mount an xterm.js terminal instance rendering the pane's ANSI buffer.
  test_type: E2E
  expected:
    ui_state: "xterm canvas/DOM rendered with correct ANSI coloring, FitAddon applied"
    url: "/panes/:id"
    side_effects: []

- id: v1.REQ-007-AC-002
  ears: >
    While in the detail view,
    the UI shall display a sidebar or drawer showing the chronologically ordered state timeline events for this pane.
  test_type: E2E
  expected:
    ui_state: "Timeline items visible displaying timestamp, from_state -> to_state, and trigger explanation"
    url: "/panes/:id"
    side_effects: []

- id: v1.REQ-007-AC-003
  ears: >
    While in MVP mode,
    the terminal input shall be read-only to prevent unauthorized execution.
  test_type: E2E
  expected:
    ui_state: "Keyboard input is disabled in xterm terminal, read-only badge displayed"
    url: "/panes/:id"
    side_effects: []
```

---

## 3. 质量自检清单

- [x] 每条 EARS 语法中系统响应均具备确定性、可测试性，无模糊主观词。
- [x] 覆盖了异常场景：tmux 宕机/未启动、非法 YAML 配置、未授权访问。
- [x] 每条 AC 均明确关联了 `test_type`（Unit / API / E2E）。
- [x] 需求编号连续（`v1.REQ-001` 至 `v1.REQ-007`），无重复或断层。

