# 数据与配置模型设计（Data Models）

> 文件路径：`docs/spec/versions/v2/data.md`
> 适用版本：v2 · 生成日期：2026-11-12
> 引用自：`docs/spec/versions/v2/design.md`
> 继承自：`docs/spec/versions/v1-finished/data.md`（保留所有基础拓扑、状态枚举与智能体规则定义）

---

## 1. 概述

v2 版本在 v1 纯内存实时模型的基础上，扩展了**轻量级本地快照持久化存储模型（History Storage Engine）**。
设计目标如下：
1. **零外部数据库依赖**：延续单二进制无外部数据库分发优势，基于结构化追加日志（JSON Lines, `.jsonl`）实现本地磁盘持久化。
2. **读写分离与元数据轻量化**：拆分全量快照（`SnapshotRecord`）与轻量列表元数据（`SnapshotMetadata`），防止历史回放查询列表时拉取过量 ANSI 输出导致网络与内存瓶颈。
3. **容量硬上限与 FIFO 自动淘汰**：单 Pane 严格限制快照记录数与存储体积，防止长期并发运行的 Agent 产生磁盘爆炸。

---

## 2. 配置模型扩展（YAML）

在 `config/agents.yaml` 中新增 `history` 与 `mobile` 配置段落：

```yaml
# 采集器基础配置（沿用 v1）
collector:
  socket_name: ""
  socket_path: ""
  poll_interval_ms: 800
  capture_tail_lines: 50
  max_buffer_bytes: 524288
  inactivity_timeout_s: 15

# v2 新增：历史输出持久化与回放配置
history:
  enabled: true                    # 是否启用历史快照持久化
  storage_dir: "data/history"      # 历史快照文件存储根目录（相对于工作区）
  max_snapshots_per_pane: 200      # 单个 Pane 最大保留历史快照数量（超额 FIFO 淘汰）
  max_bytes_per_pane: 10485760     # 单个 Pane 最大磁盘占用（10MB）
  checkpoint_interval_s: 30        # 周期性保底快照间隔（秒，输出有变动且无状态跃迁时触发）
  retention_hours: 48              # 历史快照生命周期 TTL（小时，超时文件自动清理）

# 各类型智能体匹配规则（沿用 v1 定义）
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
        tail_regex: "Tokens:|Applied edit to"
      waiting_input:
        tail_regex: "Add .* to the chat\\?|\\(Y\\)es/\\(N\\)o"
```

---

## 3. 历史快照数据模型（Go Entity）

### 3.1 历史快照元数据（`SnapshotMetadata`）
用于前端时间轴滑块刻度渲染和快照列表查询：

```go
package model

// SnapshotMetadata 历史快照轻量元数据（不包含大体积 ANSI 缓冲区）
type SnapshotMetadata struct {
	SnapshotID string `json:"snapshot_id"` // 格式：snap_{pane_id}_{timestamp}_{rand}
	PaneID     string `json:"pane_id"`     // Tmux Pane ID，如 "%0"
	Timestamp  int64  `json:"timestamp"`   // 毫秒时间戳 (Unix Milli)
	EventType  string `json:"event_type"`  // 事件类型：state_transition / checkpoint / manual
	OldState   State  `json:"old_state"`   // 跃迁前状态（如有）
	NewState   State  `json:"new_state"`   // 当前状态
	Summary    string `json:"summary"`     // 简要摘要，如 "Transitioned to waiting_input"
	ByteLength int    `json:"byte_length"` // 该快照全量 ANSI 输出字节大小
}
```

### 3.2 完整快照记录（`SnapshotRecord`）
用于指定时刻的回放渲染与持久化写入：

```go
package model

// SnapshotRecord 完整的历史快照实体
type SnapshotRecord struct {
	SnapshotMetadata
	SessionName  string   `json:"session_name"`  // Tmux 会话名
	WindowID     string   `json:"window_id"`     // Tmux 窗口 ID
	OutputBuffer string   `json:"output_buffer"` // 捕获时刻的完整终端 ANSI 字符串
	TailLines    []string `json:"tail_lines"`    // 捕获时刻末尾 N 行文本
}
```

---

## 4. 磁盘持久化结构（JSONL Append-Only）

### 4.1 目录组织与文件命名
持久化文件存储在 `data/history/` 目录下，按 Session 和 Pane 分层隔离：

```
data/history/
├── main/
│   ├── _0.jsonl        # Pane %0 的追加日志（% 被转义为 _）
│   └── _1.jsonl        # Pane %1 的追加日志
└── dev/
    └── _2.jsonl
```

### 4.2 单行存储格式
每行存储一个单行紧凑 JSON 格式的 `SnapshotRecord`：

```json
{"snapshot_id":"snap__0_1731380400000_a1b2","pane_id":"%0","timestamp":1731380400000,"event_type":"state_transition","old_state":"working","new_state":"waiting_input","summary":"State changed from working to waiting_input","byte_length":14250,"session_name":"main","window_id":"@0","output_buffer":"\u001b[32m[Agent]\u001b[0m Do you want to proceed? (y/n)","tail_lines":["Do you want to proceed? (y/n)"]}
```

### 4.3 滚动与裁剪机制（Pruning Protocol）
- **触发时机**：每次写入新快照后，若该 Pane 的记录数超过 `max_snapshots_per_pane` 或文件大小超过 `max_bytes_per_pane`。
- **裁剪算法**：读取现有记录，保留时间最新的 $N$ 条（默认 200 条），原子性重写对应 `.jsonl` 临时文件并重命名替换原文件。
