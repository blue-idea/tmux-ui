package detector

import (
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"tmux-ui/internal/model"
)

// EvaluationInput 状态研判输入参数
type EvaluationInput struct {
	PaneID            string
	IsDead            bool
	CurrentCommand    string
	AgentType         string
	TailLines         []string
	HasChanged        bool
	LastActiveAt      time.Time
	Now               time.Time
	InactivityTimeout time.Duration
	Rule              *model.AgentRule
}

// EvaluationResult 状态研判结果
type EvaluationResult struct {
	State       model.State
	TriggerRule string
}

// Engine 规则与启发式多智能体状态研判引擎
type Engine struct {
	mu           sync.RWMutex
	regexCache   map[string]*regexp.Regexp
}

// NewEngine 构造 Engine 实例
func NewEngine() *Engine {
	return &Engine{
		regexCache: make(map[string]*regexp.Regexp),
	}
}

// getCompiledRegex 获取或缓存编译正则表达式
func (e *Engine) getCompiledRegex(pattern string) (*regexp.Regexp, error) {
	e.mu.RLock()
	re, ok := e.regexCache[pattern]
	e.mu.RUnlock()
	if ok {
		return re, nil
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	// 双重校验
	if re, ok = e.regexCache[pattern]; ok {
		return re, nil
	}

	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	e.regexCache[pattern] = compiled
	return compiled, nil
}

// matchLines 检查末尾行列表中是否有任意一行命中正则
func (e *Engine) matchLines(lines []string, pattern string) bool {
	if pattern == "" || len(lines) == 0 {
		return false
	}
	re, err := e.getCompiledRegex(pattern)
	if err != nil {
		return false
	}
	for _, line := range lines {
		if re.MatchString(line) {
			return true
		}
	}
	return false
}

// isShellCommand 判定命令是否为常见纯交互 Shell
func isShellCommand(cmd string) bool {
	base := filepath.Base(cmd)
	base = strings.TrimPrefix(base, "-") // 适配 login shell 如 -zsh
	switch base {
	case "bash", "zsh", "sh", "fish", "ksh", "csh":
		return true
	default:
		return false
	}
}

// Evaluate 依据优先级决策链研判智能体状态
func (e *Engine) Evaluate(in EvaluationInput) EvaluationResult {
	// 1. 最高优先级：pane_dead == 1 直接判定为 exited
	if in.IsDead {
		return EvaluationResult{
			State:       model.StateExited,
			TriggerRule: "pane dead (pane_dead=1)",
		}
	}

	// 2. 匹配规则引擎定义的状态正则
	if in.Rule != nil && in.Rule.States != nil {
		// 2.1 waiting_input 正则匹配
		if rule, ok := in.Rule.States[model.StateWaitingInput]; ok {
			if e.matchLines(in.TailLines, rule.TailRegex) {
				return EvaluationResult{
					State:       model.StateWaitingInput,
					TriggerRule: "regex matched waiting_input",
				}
			}
		}

		// 2.2 error 正则匹配
		if rule, ok := in.Rule.States[model.StateError]; ok {
			if e.matchLines(in.TailLines, rule.TailRegex) {
				return EvaluationResult{
					State:       model.StateError,
					TriggerRule: "regex matched error",
				}
			}
		}

		// 2.3 working 正则匹配
		if rule, ok := in.Rule.States[model.StateWorking]; ok {
			if e.matchLines(in.TailLines, rule.TailRegex) {
				return EvaluationResult{
					State:       model.StateWorking,
					TriggerRule: "regex matched working",
				}
			}
		}
	}

	// 3. 启发式活跃度研判
	timeout := in.InactivityTimeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	now := in.Now
	if now.IsZero() {
		now = time.Now()
	}

	// 输出变动或处于活跃超时时间窗口内
	if in.HasChanged {
		return EvaluationResult{
			State:       model.StateWorking,
			TriggerRule: "output hash changed recently",
		}
	}
	if !in.LastActiveAt.IsZero() && now.Sub(in.LastActiveAt) < timeout {
		return EvaluationResult{
			State:       model.StateWorking,
			TriggerRule: "output active within inactivity threshold",
		}
	}

	// 4. 纯 shell 进程兜底：无子 agent 进程且输出超时
	if isShellCommand(in.CurrentCommand) && in.AgentType == "" {
		return EvaluationResult{
			State:       model.StateIdle,
			TriggerRule: "shell prompt with no running agent",
		}
	}

	// 5. 默认状态：idle
	return EvaluationResult{
		State:       model.StateIdle,
		TriggerRule: "inactivity timeout exceeded",
	}
}
