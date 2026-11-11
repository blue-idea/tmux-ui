package model

// StateMatchRule 状态匹配正则规则
type StateMatchRule struct {
	TailRegex string `yaml:"tail_regex" json:"tail_regex"`
}

// AgentRule 单个智能体的识别配置
type AgentRule struct {
	Name         string                   `yaml:"name" json:"name"`
	MatchProcess []string                 `yaml:"match_process" json:"match_process"`
	States       map[State]StateMatchRule `yaml:"states" json:"states"`
}

// CollectorConfig 采集器参数配置
type CollectorConfig struct {
	SocketName         string `yaml:"socket_name" json:"socket_name"`
	SocketPath         string `yaml:"socket_path" json:"socket_path"`
	PollIntervalMs     int    `yaml:"poll_interval_ms" json:"poll_interval_ms"`
	CaptureTailLines   int    `yaml:"capture_tail_lines" json:"capture_tail_lines"`
	MaxBufferBytes     int    `yaml:"max_buffer_bytes" json:"max_buffer_bytes"`
	InactivityTimeoutS int    `yaml:"inactivity_timeout_s" json:"inactivity_timeout_s"`
}

// HistoryConfig 历史输出持久化与回放配置
type HistoryConfig struct {
	Enabled             bool   `yaml:"enabled" json:"enabled"`
	StorageDir          string `yaml:"storage_dir" json:"storage_dir"`
	MaxSnapshotsPerPane int    `yaml:"max_snapshots_per_pane" json:"max_snapshots_per_pane"`
	MaxBytesPerPane     int64  `yaml:"max_bytes_per_pane" json:"max_bytes_per_pane"`
	CheckpointIntervalS int    `yaml:"checkpoint_interval_s" json:"checkpoint_interval_s"`
	RetentionHours      int    `yaml:"retention_hours" json:"retention_hours"`
}

// Config 全局配置结构体
type Config struct {
	Collector CollectorConfig `yaml:"collector" json:"collector"`
	History   HistoryConfig   `yaml:"history" json:"history"`
	Agents    []AgentRule     `yaml:"agents" json:"agents"`
}
