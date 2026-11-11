package model_test

import (
	"testing"

	"tmux-ui/internal/model"
)

// 验证 State 枚举合法性与字符串表示
func TestState_IsValid(t *testing.T) {
	validStates := []model.State{
		model.StateWorking,
		model.StateIdle,
		model.StateWaitingInput,
		model.StateError,
		model.StateExited,
	}

	for _, s := range validStates {
		if !s.IsValid() {
			t.Errorf("状态 %s 应该是合法的", s)
		}
	}

	invalidState := model.State("unknown_state")
	if invalidState.IsValid() {
		t.Errorf("状态 %s 应该是非法的", invalidState)
	}
}

