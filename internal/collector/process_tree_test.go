package collector_test

import (
	"context"
	"errors"
	"testing"

	"tmux-ui/internal/collector"
	"tmux-ui/internal/model"
)

// MockProcessExecutor 模拟 ps 命令执行
type MockProcessExecutor struct {
	Output []byte
	Err    error
}

func (m *MockProcessExecutor) Execute(ctx context.Context) ([]byte, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Output, nil
}

// v1.REQ-001-AC-002：解析 ps -axo pid=,ppid=,command= 输出并构建进程树
func TestProcessTree_ParseRawOutput(t *testing.T) {
	rawOutput := `
    1     0 /sbin/launchd
  100     1 /bin/zsh -l
  101   100 node /usr/local/bin/claude --project foo
  102   101 /usr/local/bin/claude
  200     1 /bin/bash
  201   200 python3 -m aider
  300     1 /bin/zsh
`
	executor := &MockProcessExecutor{Output: []byte(rawOutput)}
	tracer := collector.NewProcessTracer(collector.ProcessTracerOptions{Executor: executor})

	tree, err := tracer.CollectProcessTree(context.Background())
	if err != nil {
		t.Fatalf("收集进程树不应返回错误: %v", err)
	}

	if tree == nil {
		t.Fatal("返回的进程树不应为 nil")
	}

	// 验证进程解析数量
	procs := tree.AllProcesses()
	if len(procs) != 7 {
		t.Fatalf("期望解析出 7 个进程，实际为 %d", len(procs))
	}

	// 验证特定进程信息
	p101, exists := tree.FindProcess(101)
	if !exists {
		t.Fatal("未找到 PID 101 的进程")
	}
	if p101.PPID != 100 || p101.Command != "node /usr/local/bin/claude --project foo" {
		t.Errorf("进程 101 信息解析错误: %+v", p101)
	}
}

// v1.REQ-001-AC-002：按 Pane 根 PID 查询所有子进程
func TestProcessTree_GetPaneProcesses(t *testing.T) {
	rawOutput := `
  100     1 /bin/zsh
  101   100 node /usr/local/bin/claude
  102   101 /bin/sh -c git diff
  200     1 /bin/bash
`
	executor := &MockProcessExecutor{Output: []byte(rawOutput)}
	tracer := collector.NewProcessTracer(collector.ProcessTracerOptions{Executor: executor})

	tree, err := tracer.CollectProcessTree(context.Background())
	if err != nil {
		t.Fatalf("获取进程树失败: %v", err)
	}

	// 查询 PID 100 下的子进程
	paneProcs := tree.GetPaneProcesses(100)
	if len(paneProcs) != 3 {
		t.Fatalf("期望 PID 100 及其子孙进程共 3 个，实际为 %d", len(paneProcs))
	}

	// 查询孤立或纯 shell 无子进程的 Pane
	shellOnly := tree.GetPaneProcesses(200)
	if len(shellOnly) != 1 {
		t.Fatalf("期望 PID 200 仅包含自身 1 个进程，实际为 %d", len(shellOnly))
	}

	// 查询不存在的 PID
	none := tree.GetPaneProcesses(9999)
	if len(none) != 0 {
		t.Fatalf("不存在的 PID 应返回空切片，实际为 %d", len(none))
	}
}

// v1.REQ-001-AC-002：根据进程列表与 AgentRule 关联判定智能体类型
func TestProcessTree_MatchAgentType(t *testing.T) {
	rules := []model.AgentRule{
		{
			Name:         "claude-code",
			MatchProcess: []string{"claude", "node"},
		},
		{
			Name:         "aider",
			MatchProcess: []string{"aider", "python"},
		},
	}

	t.Run("匹配到 claude-code", func(t *testing.T) {
		procs := []model.ProcessInfo{
			{PID: 100, PPID: 1, Command: "/bin/zsh"},
			{PID: 101, PPID: 100, Command: "node /usr/local/bin/claude"},
		}
		agentType := collector.MatchAgentType(procs, rules)
		if agentType != "claude-code" {
			t.Errorf("期望匹配为 claude-code，实际为 %s", agentType)
		}
	})

	t.Run("匹配到 aider", func(t *testing.T) {
		procs := []model.ProcessInfo{
			{PID: 200, PPID: 1, Command: "/bin/bash"},
			{PID: 201, PPID: 200, Command: "python3 -m aider"},
		}
		agentType := collector.MatchAgentType(procs, rules)
		if agentType != "aider" {
			t.Errorf("期望匹配为 aider，实际为 %s", agentType)
		}
	})

	t.Run("纯 shell 未匹配到智能体", func(t *testing.T) {
		procs := []model.ProcessInfo{
			{PID: 300, PPID: 1, Command: "/bin/zsh"},
		}
		agentType := collector.MatchAgentType(procs, rules)
		if agentType != "" {
			t.Errorf("纯 shell 不应匹配到任何智能体，实际为 %s", agentType)
		}
	})
}

// 异常场景：执行 ps 出错时安全返回错误
func TestProcessTree_ExecutorError(t *testing.T) {
	executor := &MockProcessExecutor{Err: errors.New("command failed")}
	tracer := collector.NewProcessTracer(collector.ProcessTracerOptions{Executor: executor})

	tree, err := tracer.CollectProcessTree(context.Background())
	if err == nil {
		t.Fatal("执行器出错时应返回错误")
	}
	if tree != nil {
		t.Fatal("出错时返回的树应为 nil")
	}
}

// 边界场景：默认执行器与畸形行忽略
func TestProcessTree_EdgeCases(t *testing.T) {
	t.Run("默认构造函数返回可用实例", func(t *testing.T) {
		pt := collector.NewProcessTracer(collector.ProcessTracerOptions{})
		if pt == nil {
			t.Fatal("NewProcessTracer 不应返回 nil")
		}
	})

	t.Run("畸形输出行安全忽略", func(t *testing.T) {
		raw := `
  invalid
  abc def command
  100 bad_ppid command
  100 1
`
		executor := &MockProcessExecutor{Output: []byte(raw)}
		tracer := collector.NewProcessTracer(collector.ProcessTracerOptions{Executor: executor})
		tree, err := tracer.CollectProcessTree(context.Background())
		if err != nil {
			t.Fatalf("遇到畸形行不应报错: %v", err)
		}
		if len(tree.AllProcesses()) != 0 {
			t.Errorf("所有畸形行应被忽略，实际解析出: %d 个", len(tree.AllProcesses()))
		}
	})
}

