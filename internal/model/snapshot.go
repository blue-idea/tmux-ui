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

// SnapshotRecord 完整的历史快照实体
type SnapshotRecord struct {
	SnapshotMetadata
	SessionName  string   `json:"session_name"`  // Tmux 会话名
	WindowID     string   `json:"window_id"`     // Tmux 窗口 ID
	OutputBuffer string   `json:"output_buffer"` // 捕获时刻的完整终端 ANSI 字符串
	TailLines    []string `json:"tail_lines"`    // 捕获时刻末尾 N 行文本
}
