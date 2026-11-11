# 数据与配置模型设计（Data Models）

> 文件路径：`docs/spec/versions/v1/data.md`
> 适用版本：v1 · 生成日期：2026-11-11
> 引用自：`docs/spec/versions/v1/design.md`

---

## 1. 概述

本项目采用**无关系型数据库**架构，数据分为两大核心部分：
1. **静态配置数据**：由集中式 YAML 文件（`config/agents.yaml`）定义不同 Agent 的匹配规则。
2. **运行时内存模型**：由 Go 后端 `internal/store` 内存存储管理，维护当前活动的 tmux 拓扑快照、每个 pane 的终端输出缓冲区（Ring Buffer）以及状态变迁时间线。

---

## 2. 配置数据模型（YAML）

配置文件路径为 `config/agents.yaml`，其结构定义如下：

```yaml
# 全局轮询与采集配置
collector:
  socket_name: ""              # 自定义 tmux socket 名（对应 tmux -L）
  socket_path: ""              # 自定义 tmux socket 绝对路径（对应 tmux -S）
  poll_interval_ms: 800        # 拓扑轮询间隔（毫秒）
  capture_tail_lines: 50       # 用于状态判定匹配的终端末尾行数
  max_buffer_bytes: 524288     # 单个 Pane 最大缓存输出（512KB）
  inactivity_timeout_s: 15     # 输出无变动判定为 idle 的超时阈值（秒）

# 各类型智能体匹配规则
agents:
  - name: "claude-code"
    match_process: ["claude", "node"]
    states:
      working:
        tail_regex: "esc to interrupt|Thinking\\.\\.\\."
      waiting_input:
        tail_regex: "Do you want to proceed\\?|\\(y/n\\)|Allow this action\\?"
      error:
        tail_regex: "Error:|fatal:|Process exited with code [1-9]"

  - name: "codex"
    match_process: ["codex"]
    states:
      working:
        tail_regex: "Generating|Executing command"
      waiting_input:
        tail_regex: "Confirm execution\\? \\[Y/n\\]"
      error:
        tail_regex: "Failed to generate|Exception:"

  - name: "gemini-cli"
    match_process: ["gemini"]
    states:
      working:
        tail_regex: "Thinking\\.\\.\\."
      waiting_input:
        tail_regex: "Press Enter to continue|Confirm \\(yes/no\\):"

  - name: "aider"
    match_process: ["aider", "python"]
    states:
      working:
        tail_regex: "Tokens: [0-9]+|Thinking"
      waiting_input:
        tail_regex: "Apply these changes\\?|Run shell command\\?"
```

---

## 3. 运行时内存数据模型（Go）

### 3.1 状态枚举

```go
type State string

const (
    StateWorking      State = "working"       // 智能体正在思考或执行命令（绿色）
    StateIdle         State = "idle"          // 处于空闲状态，等待指令（灰色）
    StateWaitingInput State = "waiting_input" // 等待用户输入/确认授权（黄色/高亮）
    StateError        State = "error"         // 执行出错或异常终止（红色）
    StateExited       State = "exited"        // Pane 已死亡或进程已退出（暗灰色）
)
```

### 3.2 进程信息模型

```go
type ProcessInfo struct {
    PID         int           `json:"pid"`
    PPID        int           `json:"ppid"`
    Command     string        `json:"command"`
    Children    []ProcessInfo `json:"children,omitempty"`
}
```

### 3.3 Pane 基础拓扑与摘要模型（`PaneSummary`）

用于 `/api/v1/panes` 列表展示与 Web 总览卡片墙：

```go
type PaneSummary struct {
    PaneID          string    `json:"pane_id"`          // 例如 "%1"
    SessionName     string    `json:"session_name"`     // 例如 "agent-dev"
    WindowID        string    `json:"window_id"`        // 例如 "@0"
    WindowName      string    `json:"window_name"`      // 例如 "zsh"
    PanePID         int       `json:"pane_pid"`         // 根 shell PID
    CurrentCommand  string    `json:"current_command"`  // pane_current_command
    AgentType       string    `json:"agent_type"`       // 识别出的 Agent 名称，如 "claude-code"
    State           State     `json:"state"`            // 当前状态
    LastActiveAt    time.Time `json:"last_active_at"`   // 上次输出哈希变动时间
    IdleSeconds     int64     `json:"idle_seconds"`     // 空闲秒数
    IsDead          bool      `json:"is_dead"`          // pane_dead
    PreviewLines    []string  `json:"preview_lines"`    // 最近 3~5 行终端输出预览
}
```

### 3.4 Pane 详情与完整终端模型（`PaneDetail`）

用于 `/api/v1/panes/:id` 终端详情页（结合 xterm.js 渲染）：

```go
type PaneDetail struct {
    PaneSummary
    RawOutputANSI string          `json:"raw_output_ansi"` // 完整 ANSI 格式捕获输出
    Timeline      []TimelineEvent `json:"timeline"`        // 状态变迁历史（上限 100 条）
    Processes     []ProcessInfo   `json:"processes"`       // 进程树
}
```

### 3.5 状态变迁时间线事件（`TimelineEvent`）

```go
type TimelineEvent struct {
    ID          string    `json:"id"`
    PaneID      string    `json:"pane_id"`
    Timestamp   time.Time `json:"timestamp"`
    FromState   State     `json:"from_state"`
    ToState     State     `json:"to_state"`
    TriggerRule string    `json:"trigger_rule"` // 触发原因（如 "matched waiting_input regex", "hash unchanged > 15s"）
}
```

---

## 4. 状态跃迁判定矩阵与并发保障

### 4.1 状态检测优先级链

```
                  ┌───────────────────────┐
                  │ 采集 pane 元数据与输出 │
                  └───────────┬───────────┘
                              ▼
                   pane_dead == 1 ? ──────────────► 【exited】
                              │ 否
                              ▼
           匹配 YAML waiting_input 正则 ? ─────────► 【waiting_input】
                              │ 否
           匹配 YAML error 正则 ? ────────────────► 【error】
                              │ 否
           匹配 YAML working 正则 ? ──────────────► 【working】
                              │ 否
         输出 Hash 在 15s 内持续更新 ? ────────────► 【working】
                              │ 否
      当前进程为纯 shell 且无子进程 ? ──────────► 【idle】
                              │ 否
                              ▼
                     保持上一状态或置为 【idle】
```

### 4.2 线程安全与内存边界

1. **读写锁保护**：内存状态中心 `StateStore` 内部采用 `sync.RWMutex`。定时 Collector 批量加写锁写入新快照，HTTP/WS 连接并发加读锁提取视图。
2. **容量边界**：
   - 终端捕获缓冲区：单个 pane 最大保留 512KB 输出，超出使用滑窗截断头部。
   - 时间线记录：每个 pane 采用容量为 100 的环形缓冲（Ring Buffer），防止内存无界泄露。
   - 僵尸 Pane 清理：当 pane 被 tmux 销毁超过 10 分钟后，从内存 Store 中自动剔除。

