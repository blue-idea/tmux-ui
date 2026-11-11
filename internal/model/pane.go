package model

import "time"

// ProcessInfo 表示进程树节点信息
type ProcessInfo struct {
	PID      int           `json:"pid"`
	PPID     int           `json:"ppid"`
	Command  string        `json:"command"`
	Children []ProcessInfo `json:"children,omitempty"`
}

// PaneSummary 表示总览卡片墙所需的数据摘要
type PaneSummary struct {
	PaneID         string    `json:"pane_id"`         // 例如 "%1"
	SessionName    string    `json:"session_name"`    // 会话名称
	WindowID       string    `json:"window_id"`       // 例如 "@0"
	WindowName     string    `json:"window_name"`     // 窗口名称
	PanePID        int       `json:"pane_pid"`        // 根进程 PID
	CurrentCommand string    `json:"current_command"` // pane_current_command
	AgentType      string    `json:"agent_type"`      // 识别出的 Agent 名称，如 "claude-code"
	State          State     `json:"state"`           // 判定状态
	LastActiveAt   time.Time `json:"last_active_at"`  // 上次内容变更时间
	IdleSeconds    int64     `json:"idle_seconds"`    // 空闲秒数
	IsDead         bool      `json:"is_dead"`         // pane_dead
	PreviewLines   []string  `json:"preview_lines"`   // 末尾几行终端输出预览
}

// PaneDetail 表示详情页所需的完整数据
type PaneDetail struct {
	PaneSummary
	RawOutputANSI string          `json:"raw_output_ansi"` // 完整 ANSI 格式捕获输出
	Timeline      []TimelineEvent `json:"timeline"`        // 状态变迁历史（上限 100 条）
	Processes     []ProcessInfo   `json:"processes"`       // 挂载的进程树
}

// TimelineEvent 记录状态迁移时间线
type TimelineEvent struct {
	ID          string    `json:"id"`
	PaneID      string    `json:"pane_id"`
	Timestamp   time.Time `json:"timestamp"`
	FromState   State     `json:"from_state"`
	ToState     State     `json:"to_state"`
	TriggerRule string    `json:"trigger_rule"` // 触发原因
}
