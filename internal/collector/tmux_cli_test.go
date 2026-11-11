package collector_test

import (
	"context"
	"errors"
	"testing"

	"tmux-ui/internal/collector"
	"tmux-ui/internal/model"
)

// MockTmuxExecutor 用于测试隔离的命令模拟执行器
type MockTmuxExecutor struct {
	Outputs map[string][]byte
	Errors  map[string]error
}

func (m *MockTmuxExecutor) Execute(ctx context.Context, args ...string) ([]byte, error) {
	key := args[0]
	// 如果包含 -L 或 -S，取真实的子命令（如 list-panes 或 -V）
	for i, arg := range args {
		if arg == "-V" || arg == "list-panes" || arg == "capture-pane" {
			key = args[i]
			break
		}
	}

	if err, exists := m.Errors[key]; exists && err != nil {
		out := m.Outputs[key]
		return out, err
	}
	if out, exists := m.Outputs[key]; exists {
		return out, nil
	}
	return nil, errors.New("command not found in mock")
}

// v1.REQ-001-AC-004：检测 tmux -V 版本并解析 Major/Minor
func TestDetectTmuxVersion(t *testing.T) {
	testCases := []struct {
		name          string
		rawOutput     string
		expectedMajor int
		expectedMinor int
	}{
		{
			name:          "标准版本 3.4",
			rawOutput:     "tmux 3.4\n",
			expectedMajor: 3,
			expectedMinor: 4,
		},
		{
			name:          "带字符版本 3.2a",
			rawOutput:     "tmux 3.2a\n",
			expectedMajor: 3,
			expectedMinor: 2,
		},
		{
			name:          "开发版 next-3.5",
			rawOutput:     "tmux next-3.5\n",
			expectedMajor: 3,
			expectedMinor: 5,
		},
		{
			name:          "无法识别版本退回默认",
			rawOutput:     "custom-tmux-build\n",
			expectedMajor: 2,
			expectedMinor: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &MockTmuxExecutor{
				Outputs: map[string][]byte{
					"-V": []byte(tc.rawOutput),
				},
			}
			c := collector.NewTmuxCollector(collector.Options{Executor: mock})
			ver, err := c.DetectVersion(context.Background())
			if err != nil {
				t.Fatalf("检测版本不应返回错误: %v", err)
			}
			if ver.Major != tc.expectedMajor || ver.Minor != tc.expectedMinor {
				t.Errorf("期望版本 %d.%d，实际为 %d.%d", tc.expectedMajor, tc.expectedMinor, ver.Major, ver.Minor)
			}
		})
	}

	t.Run("检测版本服务不可用时返回 ErrTmuxUnavailable", func(t *testing.T) {
		mock := &MockTmuxExecutor{
			Outputs: map[string][]byte{
				"-V": []byte("failed to connect to server"),
			},
			Errors: map[string]error{
				"-V": errors.New("exit status 1"),
			},
		}
		c := collector.NewTmuxCollector(collector.Options{Executor: mock})
		_, err := c.DetectVersion(context.Background())
		if !errors.Is(err, collector.ErrTmuxUnavailable) {
			t.Errorf("期望返回 ErrTmuxUnavailable，实际为 %v", err)
		}
	})

	t.Run("检测版本普通错误", func(t *testing.T) {
		mock := &MockTmuxExecutor{
			Outputs: map[string][]byte{
				"-V": []byte("command not found"),
			},
			Errors: map[string]error{
				"-V": errors.New("exec error"),
			},
		}
		c := collector.NewTmuxCollector(collector.Options{Executor: mock})
		_, err := c.DetectVersion(context.Background())
		if err == nil || errors.Is(err, collector.ErrTmuxUnavailable) {
			t.Errorf("普通错误应包装返回普通 error，实际为 %v", err)
		}
	})
}

// 验证默认构造函数与容错
func TestNewTmuxCollector_DefaultExecutor(t *testing.T) {
	c := collector.NewTmuxCollector(collector.Options{})
	if c == nil {
		t.Fatal("NewTmuxCollector 不应返回 nil")
	}
}

// 验证 ListPanes 普通执行错误处理
func TestListPanes_GenericError(t *testing.T) {
	mock := &MockTmuxExecutor{
		Outputs: map[string][]byte{
			"list-panes": []byte("permission denied"),
		},
		Errors: map[string]error{
			"list-panes": errors.New("generic error"),
		},
	}
	c := collector.NewTmuxCollector(collector.Options{Executor: mock})
	_, err := c.ListPanes(context.Background())
	if err == nil || errors.Is(err, collector.ErrTmuxUnavailable) {
		t.Errorf("应返回常规执行错误，实际为: %v", err)
	}
}


// v1.REQ-001-AC-005：构建 tmux 命令时支持 CLI 参数与配置文件的 Socket 优先级
func TestBuildTmuxArgs_SocketPrecedence(t *testing.T) {
	t.Run("CLI 参数优先级高于配置", func(t *testing.T) {
		opts := collector.Options{
			CLISocketName: "cli_sock",
			Config: model.CollectorConfig{
				SocketName: "config_sock",
			},
		}
		c := collector.NewTmuxCollector(opts)
		args := c.BuildArgs("list-panes")
		if len(args) < 3 || args[0] != "-L" || args[1] != "cli_sock" || args[2] != "list-panes" {
			t.Errorf("CLI socket 应覆盖配置，实际生成参数: %+v", args)
		}
	})

	t.Run("配置指定的 SocketPath 转换为 -S", func(t *testing.T) {
		opts := collector.Options{
			Config: model.CollectorConfig{
				SocketPath: "/tmp/custom.sock",
			},
		}
		c := collector.NewTmuxCollector(opts)
		args := c.BuildArgs("list-panes")
		if len(args) < 3 || args[0] != "-S" || args[1] != "/tmp/custom.sock" || args[2] != "list-panes" {
			t.Errorf("应生成 -S /tmp/custom.sock 参数，实际生成: %+v", args)
		}
	})
}

// v1.REQ-001-AC-001：执行 list-panes 并准确反序列化为 PaneSummary 列表
func TestListPanes_Success(t *testing.T) {
	rawOutput := "agent-dev\t@0\tclaude-win\t%1\t12345\tnode\t0\n" +
		"agent-dev\t@1\tdead-win\t%2\t12346\tzsh\t1\n"

	mock := &MockTmuxExecutor{
		Outputs: map[string][]byte{
			"list-panes": []byte(rawOutput),
		},
	}

	c := collector.NewTmuxCollector(collector.Options{Executor: mock})
	panes, err := c.ListPanes(context.Background())
	if err != nil {
		t.Fatalf("ListPanes 不应报错: %v", err)
	}

	if len(panes) != 2 {
		t.Fatalf("期望解析出 2 个 pane，实际为 %d", len(panes))
	}

	p1 := panes[0]
	if p1.SessionName != "agent-dev" || p1.WindowID != "@0" || p1.WindowName != "claude-win" ||
		p1.PaneID != "%1" || p1.PanePID != 12345 || p1.CurrentCommand != "node" || p1.IsDead != false {
		t.Errorf("Pane 1 解析字段不符: %+v", p1)
	}

	p2 := panes[1]
	if p2.PaneID != "%2" || p2.IsDead != true {
		t.Errorf("Pane 2 死亡标记解析错误: %+v", p2)
	}
}

// v1.REQ-001-AC-003：tmux 未启动或 socket 不可达时返回明确类型错误
func TestListPanes_ServerUnavailable(t *testing.T) {
	mock := &MockTmuxExecutor{
		Outputs: map[string][]byte{
			"list-panes": []byte("no server running on /tmp/tmux-501/default\n"),
		},
		Errors: map[string]error{
			"list-panes": errors.New("exit status 1"),
		},
	}

	c := collector.NewTmuxCollector(collector.Options{Executor: mock})
	panes, err := c.ListPanes(context.Background())
	if !errors.Is(err, collector.ErrTmuxUnavailable) {
		t.Errorf("期望返回 ErrTmuxUnavailable 错误，实际为: %v", err)
	}
	if panes != nil {
		t.Errorf("异常时返回的 panes 应为 nil，实际为: %+v", panes)
	}
}
