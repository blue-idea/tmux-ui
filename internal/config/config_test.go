package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"tmux-ui/internal/config"
	"tmux-ui/internal/model"
)

// v1.REQ-004-AC-001：正常加载 YAML 配置文件并解析 Agent 匹配模式与状态正则
func TestLoadConfig_FromValidYAML(t *testing.T) {
	yamlContent := `
collector:
  socket_name: "test_sock"
  socket_path: "/tmp/tmux_test.sock"
  poll_interval_ms: 500
  capture_tail_lines: 30
  max_buffer_bytes: 262144
  inactivity_timeout_s: 10

agents:
  - name: "claude-code"
    match_process: ["claude", "node"]
    states:
      working:
        tail_regex: "esc to interrupt"
      waiting_input:
        tail_regex: "Do you want to proceed\\?"
      error:
        tail_regex: "Error:"
`
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "agents.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("创建临时测试配置文件失败: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("加载有效配置不应返回错误: %v", err)
	}

	if cfg == nil {
		t.Fatal("返回的配置不应为 nil")
	}

	// 验证 Collector 配置
	if cfg.Collector.SocketName != "test_sock" {
		t.Errorf("期望 SocketName 为 'test_sock'，实际为 '%s'", cfg.Collector.SocketName)
	}
	if cfg.Collector.PollIntervalMs != 500 {
		t.Errorf("期望 PollIntervalMs 为 500，实际为 %d", cfg.Collector.PollIntervalMs)
	}
	if cfg.Collector.CaptureTailLines != 30 {
		t.Errorf("期望 CaptureTailLines 为 30，实际为 %d", cfg.Collector.CaptureTailLines)
	}

	// 验证 Agent 规则解析
	if len(cfg.Agents) != 1 {
		t.Fatalf("期望解析出 1 个 agent，实际为 %d", len(cfg.Agents))
	}
	agent := cfg.Agents[0]
	if agent.Name != "claude-code" {
		t.Errorf("期望 Agent 名称为 claude-code，实际为 %s", agent.Name)
	}
	if len(agent.MatchProcess) != 2 || agent.MatchProcess[0] != "claude" {
		t.Errorf("进程匹配列表解析错误: %+v", agent.MatchProcess)
	}
	if agent.States[model.StateWorking].TailRegex != "esc to interrupt" {
		t.Errorf("working 状态正则解析错误: %s", agent.States[model.StateWorking].TailRegex)
	}
	if agent.States[model.StateWaitingInput].TailRegex != "Do you want to proceed\\?" {
		t.Errorf("waiting_input 状态正则解析错误: %s", agent.States[model.StateWaitingInput].TailRegex)
	}
}

// v1.REQ-004-AC-002：文件不存在时安全降级至内置默认配置
func TestLoadConfig_FileNotFound_Fallback(t *testing.T) {
	nonExistentPath := filepath.Join(t.TempDir(), "non_existent_agents.yaml")

	cfg, err := config.Load(nonExistentPath)
	if err != nil {
		t.Fatalf("文件不存在时应降级而不是返回致命错误: %v", err)
	}
	if cfg == nil {
		t.Fatal("降级配置不应为 nil")
	}

	// 默认配置应具备通用 agent 规则
	if len(cfg.Agents) == 0 {
		t.Error("内置默认配置应至少包含 1 个默认智能体识别规则")
	}

	foundClaude := false
	for _, a := range cfg.Agents {
		if a.Name == "claude-code" {
			foundClaude = true
			break
		}
	}
	if !foundClaude {
		t.Error("内置默认配置中未找到 claude-code 规则")
	}
}

// v1.REQ-004-AC-002：YAML 语法损坏时安全降级至内置默认配置
func TestLoadConfig_InvalidYAML_Fallback(t *testing.T) {
	tempDir := t.TempDir()
	brokenPath := filepath.Join(tempDir, "broken.yaml")
	if err := os.WriteFile(brokenPath, []byte("broken: [yaml: :::"), 0644); err != nil {
		t.Fatalf("创建损坏 YAML 失败: %v", err)
	}

	cfg, err := config.Load(brokenPath)
	if err != nil {
		t.Fatalf("损坏 YAML 应触发降级而不是抛出致命错误: %v", err)
	}
	if cfg == nil {
		t.Fatal("降级配置不应为 nil")
	}
	if len(cfg.Agents) == 0 {
		t.Error("降级配置应包含默认智能体规则")
	}
}

// v1.REQ-004-AC-001：YAML 中 Collector 参数缺省时自动补充默认值
func TestLoadConfig_ZeroValuesDefaultBackfill(t *testing.T) {
	tempDir := t.TempDir()
	emptyYamlPath := filepath.Join(tempDir, "empty.yaml")
	if err := os.WriteFile(emptyYamlPath, []byte("collector: {}\n"), 0644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	cfg, err := config.Load(emptyYamlPath)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	if cfg.Collector.PollIntervalMs != 800 {
		t.Errorf("期望默认 PollIntervalMs 为 800，实际为 %d", cfg.Collector.PollIntervalMs)
	}
	if cfg.Collector.CaptureTailLines != 50 {
		t.Errorf("期望默认 CaptureTailLines 为 50，实际为 %d", cfg.Collector.CaptureTailLines)
	}
	if cfg.Collector.MaxBufferBytes != 524288 {
		t.Errorf("期望默认 MaxBufferBytes 为 524288，实际为 %d", cfg.Collector.MaxBufferBytes)
	}
	if cfg.Collector.InactivityTimeoutS != 15 {
		t.Errorf("期望默认 InactivityTimeoutS 为 15，实际为 %d", cfg.Collector.InactivityTimeoutS)
	}
	if len(cfg.Agents) == 0 {
		t.Error("未提供 agents 时应自动回填内置 agents 规则")
	}
}

