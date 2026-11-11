package collector

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"tmux-ui/internal/model"
)

// ProcessExecutor 抽象获取系统进程列表的执行器接口
type ProcessExecutor interface {
	Execute(ctx context.Context) ([]byte, error)
}

// DefaultProcessExecutor 真实执行系统 ps 命令
type DefaultProcessExecutor struct{}

func (e *DefaultProcessExecutor) Execute(ctx context.Context) ([]byte, error) {
	// 使用 macOS (BSD) 与 Linux (procps) 兼容的格式参数
	cmd := exec.CommandContext(ctx, "ps", "-axo", "pid=,ppid=,command=")
	return cmd.CombinedOutput()
}

// ProcessTree 表示根据 PID 与 PPID 关系的进程树索引
type ProcessTree struct {
	processes map[int]model.ProcessInfo
	children  map[int][]int // 父 PID -> 子 PID 列表
}

// AllProcesses 返回当前解析到的所有扁平进程列表
func (t *ProcessTree) AllProcesses() []model.ProcessInfo {
	list := make([]model.ProcessInfo, 0, len(t.processes))
	for _, p := range t.processes {
		list = append(list, p)
	}
	return list
}

// FindProcess 依据 PID 查找单个进程信息
func (t *ProcessTree) FindProcess(pid int) (model.ProcessInfo, bool) {
	p, ok := t.processes[pid]
	return p, ok
}

// GetPaneProcesses 递归查找并返回指定 pane 根 PID 下的所有进程（含自身与子孙进程）
func (t *ProcessTree) GetPaneProcesses(panePID int) []model.ProcessInfo {
	root, exists := t.processes[panePID]
	if !exists {
		return []model.ProcessInfo{}
	}

	var results []model.ProcessInfo
	var traverse func(current model.ProcessInfo)

	traverse = func(current model.ProcessInfo) {
		results = append(results, current)
		for _, childPID := range t.children[current.PID] {
			if child, ok := t.processes[childPID]; ok {
				traverse(child)
			}
		}
	}

	traverse(root)
	return results
}

// ProcessTracerOptions 进程追踪器的配置选项
type ProcessTracerOptions struct {
	Executor ProcessExecutor
}

// ProcessTracer 负责调用 ps 并解析系统进程树
type ProcessTracer struct {
	executor ProcessExecutor
}

// NewProcessTracer 构造 ProcessTracer 实例
func NewProcessTracer(opts ProcessTracerOptions) *ProcessTracer {
	exec := opts.Executor
	if exec == nil {
		exec = &DefaultProcessExecutor{}
	}
	return &ProcessTracer{executor: exec}
}

// CollectProcessTree 调用 ps 命令并构建 ProcessTree
func (pt *ProcessTracer) CollectProcessTree(ctx context.Context) (*ProcessTree, error) {
	out, err := pt.executor.Execute(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute ps command: %w", err)
	}

	lines := bytes.Split(out, []byte("\n"))
	procs := make(map[int]model.ProcessInfo)
	children := make(map[int][]int)

	for _, line := range lines {
		trimmed := strings.TrimSpace(string(line))
		if trimmed == "" {
			continue
		}

		fields := strings.Fields(trimmed)
		if len(fields) < 3 {
			continue
		}

		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}

		ppid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}

		// 提取从第 3 个字段开始的完整命令行
		command := strings.TrimSpace(trimmed[len(fields[0]):])
		command = strings.TrimSpace(command[len(fields[1]):])

		p := model.ProcessInfo{
			PID:     pid,
			PPID:    ppid,
			Command: command,
		}

		procs[pid] = p
		children[ppid] = append(children[ppid], pid)
	}

	return &ProcessTree{
		processes: procs,
		children:  children,
	}, nil
}

// MatchAgentType 检查进程列表是否命中某种智能体的进程特征
func MatchAgentType(procs []model.ProcessInfo, rules []model.AgentRule) string {
	for _, rule := range rules {
		for _, targetProc := range rule.MatchProcess {
			target := strings.ToLower(targetProc)
			for _, p := range procs {
				cmdLower := strings.ToLower(p.Command)
				// 判定 command 包含目标进程名（如 "node /usr/local/bin/claude" 命中 "claude"）
				if strings.Contains(cmdLower, target) {
					return rule.Name
				}
			}
		}
	}
	return ""
}
