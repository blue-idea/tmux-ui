package cli_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"tmux-ui/internal/cli"
	"tmux-ui/internal/model"
)

// v3.REQ-004-AC-001: 终端 ASCII 表格格式化
func TestFormatStatusTable(t *testing.T) {
	t.Run("空列表时输出友好提示", func(t *testing.T) {
		out := cli.FormatStatusTable(nil)
		if !strings.Contains(out, "No active tmux panes detected") {
			t.Errorf("期望包含空列表提示，实际输出为: %s", out)
		}
	})

	t.Run("包含活动 Pane 时输出对齐表格及表头", func(t *testing.T) {
		panes := []model.PaneSummary{
			{
				PaneID:         "%0",
				SessionName:    "dev",
				WindowName:     "editor",
				AgentType:      "claude-code",
				State:          model.StateWorking,
				CurrentCommand: "node",
				IdleSeconds:    5,
				PreviewLines:   []string{"Thinking..."},
			},
			{
				PaneID:         "%1",
				SessionName:    "prod",
				WindowName:     "bash",
				AgentType:      "aider",
				State:          model.StateWaitingInput,
				CurrentCommand: "python",
				IdleSeconds:    20,
				PreviewLines:   []string{"Apply these changes? (y/n)"},
			},
		}

		out := cli.FormatStatusTable(panes)

		// 验证表头
		for _, header := range []string{"PANE ID", "SESSION", "AGENT", "STATE", "COMMAND", "IDLE", "PREVIEW"} {
			if !strings.Contains(out, header) {
				t.Errorf("表格输出缺失表头: %s", header)
			}
		}

		// 验证数据行
		if !strings.Contains(out, "%0") || !strings.Contains(out, "claude-code") || !strings.Contains(out, "WORKING") {
			t.Errorf("表格输出缺失 pane %%0 数据行: %s", out)
		}
		if !strings.Contains(out, "%1") || !strings.Contains(out, "aider") || !strings.Contains(out, "WAITING_INPUT") {
			t.Errorf("表格输出缺失 pane %%1 数据行: %s", out)
		}
	})
}

// v3.REQ-004-AC-002: JSON 格式化输出
func TestFormatStatusJSON(t *testing.T) {
	panes := []model.PaneSummary{
		{
			PaneID:      "%0",
			SessionName: "dev",
			AgentType:   "claude-code",
			State:       model.StateWorking,
		},
	}

	jsonStr, err := cli.FormatStatusJSON(panes)
	if err != nil {
		t.Fatalf("FormatStatusJSON 失败: %v", err)
	}

	var parsed []model.PaneSummary
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Fatalf("JSON 反序列化失败: %v, 原始输出: %s", err, jsonStr)
	}

	if len(parsed) != 1 || parsed[0].PaneID != "%0" || parsed[0].AgentType != "claude-code" {
		t.Errorf("JSON 反序列化结果不符: %+v", parsed)
	}
}

// v3.REQ-004-AC-003: 远程服务端状态拉取与 Token 校验
func TestFetchStatusFromServer(t *testing.T) {
	expectedToken := "test_cli_secret"
	mockPanes := []model.PaneSummary{
		{
			PaneID:      "%2",
			SessionName: "remote",
			State:       model.StateIdle,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+expectedToken {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "unauthorized"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"message": "success",
			"data":    mockPanes,
		})
	}))
	defer ts.Close()

	t.Run("有效 Token 获取成功", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		panes, err := cli.FetchStatusFromServer(ctx, ts.URL, expectedToken)
		if err != nil {
			t.Fatalf("获取服务端状态失败: %v", err)
		}
		if len(panes) != 1 || panes[0].PaneID != "%2" {
			t.Errorf("获取数据不符合预期: %+v", panes)
		}
	})

	t.Run("错误 Token 时返回鉴权失败", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		_, err := cli.FetchStatusFromServer(ctx, ts.URL, "wrong_token")
		if err == nil {
			t.Fatalf("期望鉴权失败报错，但未报错")
		}
		if !strings.Contains(err.Error(), "401") && !strings.Contains(err.Error(), "unauthorized") {
			t.Errorf("错误信息不符合预期: %v", err)
		}
	})
}

