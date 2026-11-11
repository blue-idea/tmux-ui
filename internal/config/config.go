package config

import (
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
	"tmux-ui/internal/model"
)

// HistoryConfig 历史持久化配置类型别名
type HistoryConfig = model.HistoryConfig

// DefaultConfig 返回内置的默认全局配置
func DefaultConfig() *model.Config {
	return &model.Config{
		Collector: model.CollectorConfig{
			SocketName:         "",
			SocketPath:         "",
			PollIntervalMs:     800,
			CaptureTailLines:   50,
			MaxBufferBytes:     524288,
			InactivityTimeoutS: 15,
		},
		History: model.HistoryConfig{
			Enabled:             true,
			StorageDir:          "data/history",
			MaxSnapshotsPerPane: 200,
			MaxBytesPerPane:     10485760, // 10MB
			CheckpointIntervalS: 30,
			RetentionHours:      48,
		},
		Agents: []model.AgentRule{
			{
				Name:         "claude-code",
				MatchProcess: []string{"claude", "node"},
				States: map[model.State]model.StateMatchRule{
					model.StateWorking: {
						TailRegex: `esc to interrupt|Thinking\.\.\.`,
					},
					model.StateWaitingInput: {
						TailRegex: `Do you want to proceed\?|\(y/n\)|Allow this action\?`,
					},
					model.StateError: {
						TailRegex: `Error:|fatal:|Process exited with code [1-9]`,
					},
				},
			},
			{
				Name:         "codex",
				MatchProcess: []string{"codex"},
				States: map[model.State]model.StateMatchRule{
					model.StateWorking: {
						TailRegex: `Generating|Executing command`,
					},
					model.StateWaitingInput: {
						TailRegex: `Confirm execution\? \[Y/n\]`,
					},
					model.StateError: {
						TailRegex: `Failed to generate|Exception:`,
					},
				},
			},
			{
				Name:         "gemini-cli",
				MatchProcess: []string{"gemini"},
				States: map[model.State]model.StateMatchRule{
					model.StateWorking: {
						TailRegex: `Thinking\.\.\.`,
					},
					model.StateWaitingInput: {
						TailRegex: `Press Enter to continue|Confirm \(yes/no\):`,
					},
				},
			},
			{
				Name:         "aider",
				MatchProcess: []string{"aider", "python"},
				States: map[model.State]model.StateMatchRule{
					model.StateWorking: {
						TailRegex: `Tokens: [0-9]+|Thinking`,
					},
					model.StateWaitingInput: {
						TailRegex: `Apply these changes\?|Run shell command\?`,
					},
				},
			},
		},
	}
}

// Load 读取并解析指定路径的 YAML 配置文件，文件缺失或损坏时降级至内置配置
func Load(path string) (*model.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		slog.Warn("Configuration file not found, falling back to builtin defaults", "path", path, "error", err.Error())
		return DefaultConfig(), nil
	}

	var cfg model.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		slog.Warn("Failed to parse YAML configuration, falling back to builtin defaults", "path", path, "error", err.Error())
		return DefaultConfig(), nil
	}

	// 补全默认收集器数值
	if cfg.Collector.PollIntervalMs <= 0 {
		cfg.Collector.PollIntervalMs = 800
	}
	if cfg.Collector.CaptureTailLines <= 0 {
		cfg.Collector.CaptureTailLines = 50
	}
	if cfg.Collector.MaxBufferBytes <= 0 {
		cfg.Collector.MaxBufferBytes = 524288
	}
	if cfg.Collector.InactivityTimeoutS <= 0 {
		cfg.Collector.InactivityTimeoutS = 15
	}

	// 补全默认历史存储数值
	if cfg.History.StorageDir == "" {
		cfg.History.StorageDir = "data/history"
	}
	if cfg.History.MaxSnapshotsPerPane <= 0 {
		cfg.History.MaxSnapshotsPerPane = 200
	}
	if cfg.History.MaxBytesPerPane <= 0 {
		cfg.History.MaxBytesPerPane = 10485760
	}
	if cfg.History.CheckpointIntervalS <= 0 {
		cfg.History.CheckpointIntervalS = 30
	}
	if cfg.History.RetentionHours <= 0 {
		cfg.History.RetentionHours = 48
	}

	// 如果解析后未定义任何 agents，补充内置默认 agent 规则
	if len(cfg.Agents) == 0 {
		cfg.Agents = DefaultConfig().Agents
	}

	return &cfg, nil
}
