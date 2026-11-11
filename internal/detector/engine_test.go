package detector_test

import (
	"testing"
	"time"

	"tmux-ui/internal/detector"
	"tmux-ui/internal/model"
)

// v1.REQ-003-AC-001：pane_dead 为 1 时最高优先级判定为 exited
func TestDetector_PriorityDeadExited(t *testing.T) {
	d := detector.NewEngine()
	rule := &model.AgentRule{
		Name: "claude-code",
		States: map[model.State]model.StateMatchRule{
			model.StateWorking: {TailRegex: "esc to interrupt"},
		},
	}

	input := detector.EvaluationInput{
		PaneID:    "%1",
		IsDead:    true, // 进程已死
		TailLines: []string{"esc to interrupt"},
		Rule:      rule,
	}

	res := d.Evaluate(input)
	if res.State != model.StateExited {
		t.Errorf("期望状态为 exited，实际判定为 %s", res.State)
	}
}

// v1.REQ-003-AC-002：尾部文本匹配 waiting_input 正则判定为 waiting_input
func TestDetector_WaitingInputRegex(t *testing.T) {
	d := detector.NewEngine()
	rule := &model.AgentRule{
		Name: "claude-code",
		States: map[model.State]model.StateMatchRule{
			model.StateWorking:      {TailRegex: "esc to interrupt"},
			model.StateWaitingInput: {TailRegex: `Do you want to proceed\?|\(y/n\)`},
			model.StateError:        {TailRegex: `Error:|fatal:`},
		},
	}

	testCases := []struct {
		name      string
		tailLines []string
	}{
		{
			name:      "匹配 Do you want to proceed?",
			tailLines: []string{"thinking...", "Do you want to proceed? [y/N]"},
		},
		{
			name:      "匹配 (y/n)",
			tailLines: []string{"Apply git patch? (y/n)"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := detector.EvaluationInput{
				PaneID:    "%1",
				IsDead:    false,
				TailLines: tc.tailLines,
				Rule:      rule,
			}
			res := d.Evaluate(input)
			if res.State != model.StateWaitingInput {
				t.Errorf("期望判定为 waiting_input，实际为 %s (reason: %s)", res.State, res.TriggerRule)
			}
		})
	}
}

// v1.REQ-003-AC-003：尾部匹配 working 正则或输出在超时阈值内变动判定为 working
func TestDetector_WorkingEvaluation(t *testing.T) {
	d := detector.NewEngine()
	rule := &model.AgentRule{
		Name: "claude-code",
		States: map[model.State]model.StateMatchRule{
			model.StateWorking: {TailRegex: `esc to interrupt|Thinking\.\.\.`},
		},
	}

	t.Run("匹配 working 正则判定为 working", func(t *testing.T) {
		input := detector.EvaluationInput{
			PaneID:    "%1",
			IsDead:    false,
			TailLines: []string{"Thinking... esc to interrupt"},
			Rule:      rule,
		}
		res := d.Evaluate(input)
		if res.State != model.StateWorking {
			t.Errorf("期望判定为 working，实际为 %s", res.State)
		}
	})

	t.Run("输出发生变动且未超时判定为 working", func(t *testing.T) {
		now := time.Now()
		input := detector.EvaluationInput{
			PaneID:            "%1",
			IsDead:            false,
			TailLines:         []string{"arbitrary output"},
			HasChanged:        true,
			LastActiveAt:      now,
			Now:               now,
			InactivityTimeout: 15 * time.Second,
			Rule:              nil,
		}
		res := d.Evaluate(input)
		if res.State != model.StateWorking {
			t.Errorf("哈希刚变动的活跃 pane 应判定为 working，实际为 %s", res.State)
		}
	})
}

// v1.REQ-003-AC-004：当前进程为纯 shell 且无智能体子进程判定为 idle
func TestDetector_ShellIdle(t *testing.T) {
	d := detector.NewEngine()

	shellCommands := []string{"/bin/zsh", "bash", "-zsh", "sh"}
	for _, cmd := range shellCommands {
		t.Run("纯 shell "+cmd, func(t *testing.T) {
			input := detector.EvaluationInput{
				PaneID:            "%1",
				IsDead:            false,
				CurrentCommand:    cmd,
				AgentType:         "", // 无智能体子进程
				TailLines:         []string{"user@host:~$ "},
				InactivityTimeout: 15 * time.Second,
			}
			res := d.Evaluate(input)
			if res.State != model.StateIdle {
				t.Errorf("纯 shell 期望判定为 idle，实际为 %s", res.State)
			}
		})
	}
}

// 错误状态正则判定
func TestDetector_ErrorRegex(t *testing.T) {
	d := detector.NewEngine()
	rule := &model.AgentRule{
		Name: "claude-code",
		States: map[model.State]model.StateMatchRule{
			model.StateError: {TailRegex: `fatal:|Process exited with code [1-9]`},
		},
	}

	input := detector.EvaluationInput{
		PaneID:    "%1",
		IsDead:    false,
		TailLines: []string{"fatal: connection refused by server"},
		Rule:      rule,
	}

	res := d.Evaluate(input)
	if res.State != model.StateError {
		t.Errorf("期望判定为 error，实际为 %s", res.State)
	}
}

// 边界场景：正则编译错误、空行及超时兜底
func TestDetector_EdgeCases(t *testing.T) {
	d := detector.NewEngine()

	t.Run("非法正则安全忽略不崩溃", func(t *testing.T) {
		rule := &model.AgentRule{
			Name: "bad-rule",
			States: map[model.State]model.StateMatchRule{
				model.StateWaitingInput: {TailRegex: `[invalid(`},
			},
		}
		input := detector.EvaluationInput{
			PaneID:    "%1",
			TailLines: []string{"some text"},
			Rule:      rule,
		}
		res := d.Evaluate(input)
		if res.State == model.StateWaitingInput {
			t.Error("非法正则不应匹配成功")
		}
	})

	t.Run("空末尾行列表", func(t *testing.T) {
		rule := &model.AgentRule{
			Name: "claude",
			States: map[model.State]model.StateMatchRule{
				model.StateWorking: {TailRegex: `working`},
			},
		}
		input := detector.EvaluationInput{
			PaneID:    "%1",
			TailLines: []string{},
			Rule:      rule,
		}
		res := d.Evaluate(input)
		if res.State == model.StateWorking {
			t.Error("空末尾行不应判定为 working")
		}
	})

	t.Run("非 Shell 进程但超时未活动退回 Idle", func(t *testing.T) {
		now := time.Now()
		input := detector.EvaluationInput{
			PaneID:            "%1",
			CurrentCommand:    "python3 server.py",
			AgentType:         "",
			LastActiveAt:      now.Add(-30 * time.Second),
			Now:               now,
			InactivityTimeout: 15 * time.Second,
		}
		res := d.Evaluate(input)
		if res.State != model.StateIdle {
			t.Errorf("非活跃进程应超时退回 idle，实际为 %s", res.State)
		}
	})
}

