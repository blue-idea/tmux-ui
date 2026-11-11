package collector

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"tmux-ui/internal/model"
)

// ErrTmuxUnavailable 表示 tmux 服务未启动或 socket 不可访问
var ErrTmuxUnavailable = errors.New("tmux server unavailable")

// TmuxExecutor 抽象命令执行接口以便于测试注入
type TmuxExecutor interface {
	Execute(ctx context.Context, args ...string) ([]byte, error)
}

// DefaultTmuxExecutor 真实调用系统的 tmux 命令行执行器
type DefaultTmuxExecutor struct{}

func (e *DefaultTmuxExecutor) Execute(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "tmux", args...)
	// 确保注入 UTF-8 语言环境变量，防止 tmux 在 C locale 下将 format 中的制表符等非字母字符转义为下划线
	env := os.Environ()
	hasLC := false
	hasLang := false
	for _, kv := range env {
		if strings.HasPrefix(kv, "LC_ALL=") {
			hasLC = true
		}
		if strings.HasPrefix(kv, "LANG=") {
			hasLang = true
		}
	}
	if !hasLC {
		env = append(env, "LC_ALL=en_US.UTF-8")
	}
	if !hasLang {
		env = append(env, "LANG=en_US.UTF-8")
	}
	cmd.Env = env
	return cmd.CombinedOutput()
}

// TmuxVersion 表示解析出的 tmux 主次版本号
type TmuxVersion struct {
	Major int
	Minor int
	Raw   string
}

// Options 初始化 TmuxCollector 的参数选项
type Options struct {
	Executor      TmuxExecutor
	Config        model.CollectorConfig
	CLISocketName string
	CLISocketPath string
}

// TmuxCollector 负责调用 tmux CLI 进行拓扑和终端数据采集
type TmuxCollector struct {
	executor TmuxExecutor
	config   model.CollectorConfig
	cliSockN string
	cliSockP string
}

// NewTmuxCollector 实例化 TmuxCollector
func NewTmuxCollector(opts Options) *TmuxCollector {
	exec := opts.Executor
	if exec == nil {
		exec = &DefaultTmuxExecutor{}
	}
	return &TmuxCollector{
		executor: exec,
		config:   opts.Config,
		cliSockN: opts.CLISocketName,
		cliSockP: opts.CLISocketPath,
	}
}

// BuildArgs 构建带有前置 -L/-S socket 选项的完整 tmux 参数序列
func (c *TmuxCollector) BuildArgs(subCmd string, extra ...string) []string {
	var args []string

	// 命令行参数优先于配置文件：SocketPath 优先级高于 SocketName
	sockPath := c.cliSockP
	if sockPath == "" {
		sockPath = c.config.SocketPath
	}

	sockName := c.cliSockN
	if sockName == "" {
		sockName = c.config.SocketName
	}

	if sockPath != "" {
		args = append(args, "-S", sockPath)
	} else if sockName != "" {
		args = append(args, "-L", sockName)
	}

	args = append(args, subCmd)
	args = append(args, extra...)
	return args
}

var versionRegex = regexp.MustCompile(`tmux\s+(?:next-)?(\d+)\.(\d+)`)

// DetectVersion 执行 tmux -V 并解析版本信息
func (c *TmuxCollector) DetectVersion(ctx context.Context) (TmuxVersion, error) {
	out, err := c.executor.Execute(ctx, "-V")
	if err != nil {
		if isServerUnavailable(out, err) {
			return TmuxVersion{}, ErrTmuxUnavailable
		}
		return TmuxVersion{}, fmt.Errorf("failed to execute tmux -V: %w", err)
	}

	raw := strings.TrimSpace(string(out))
	matches := versionRegex.FindStringSubmatch(raw)
	if len(matches) < 3 {
		return TmuxVersion{Raw: raw, Major: 2, Minor: 0}, nil
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	return TmuxVersion{
		Major: major,
		Minor: minor,
		Raw:   raw,
	}, nil
}

// ListPanes 执行 tmux list-panes -a -F 并解析所有活动的 Pane
func (c *TmuxCollector) ListPanes(ctx context.Context) ([]model.PaneSummary, error) {
	// 格式字段依次为：session_name, window_id, window_name, pane_id, pane_pid, pane_current_command, pane_dead
	format := "#{session_name}\t#{window_id}\t#{window_name}\t#{pane_id}\t#{pane_pid}\t#{pane_current_command}\t#{pane_dead}"
	args := c.BuildArgs("list-panes", "-a", "-F", format)

	out, err := c.executor.Execute(ctx, args...)
	if err != nil {
		if isServerUnavailable(out, err) {
			return nil, ErrTmuxUnavailable
		}
		return nil, fmt.Errorf("tmux list-panes failed: %w", err)
	}

	lines := bytes.Split(out, []byte("\n"))
	var panes []model.PaneSummary

	for _, line := range lines {
		trimmed := strings.TrimSpace(string(line))
		if trimmed == "" {
			continue
		}

		parts := strings.Split(trimmed, "\t")
		if len(parts) < 7 {
			continue
		}

		pid, _ := strconv.Atoi(parts[4])
		isDead := parts[6] == "1"

		panes = append(panes, model.PaneSummary{
			SessionName:    parts[0],
			WindowID:       parts[1],
			WindowName:     parts[2],
			PaneID:         parts[3],
			PanePID:        pid,
			CurrentCommand: parts[5],
			IsDead:         isDead,
			State:          model.StateIdle, // 默认初始状态，后续由 Detector 研判
		})
	}

	return panes, nil
}

// isServerUnavailable 判定 tmux 命令报错是否由于 server 未启动或 socket 不可达造成
func isServerUnavailable(out []byte, err error) bool {
	if err == nil {
		return false
	}
	outStr := strings.ToLower(string(out))
	if strings.Contains(outStr, "no server running") ||
		strings.Contains(outStr, "failed to connect to server") ||
		strings.Contains(outStr, "error connecting to") {
		return true
	}
	return false
}
