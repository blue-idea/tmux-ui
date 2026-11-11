package model

// StateMatchRule 状态匹配正则规则
type StateMatchRule struct {
	TailRegex string `yaml:"tail_regex" json:"tail_regex"`
}

// AgentRule 单个智能体的识别配置
type AgentRule struct {
	Name         string                    `yaml:"name" json:"name"`
	MatchProcess []string                  `yaml:"match_process" json:"match_process"`
	States       map[State]StateMatchRule  `yaml:"states" json:"states"`
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

// Config 全局配置结构体
type Config struct {
	Collector CollectorConfig `yaml:"collector" json:"collector"`
	Agents    []AgentRule     `yaml:"agents" json:"agents"`
}

