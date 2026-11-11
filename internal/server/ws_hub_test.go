package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"tmux-ui/internal/model"
	"tmux-ui/internal/server"
	"tmux-ui/internal/store"
)

// 创建测试用包含 WebSocket 的 HTTP Server
func setupTestWSServer(token string) (*httptest.Server, *store.MemoryStore) {
	memStore := store.NewMemoryStore()
	cfg := &model.Config{}
	srv := server.NewHTTPServer(server.HTTPServerOptions{
		Store:       memStore,
		Config:      cfg,
		AuthToken:   token,
		TmuxVersion: "tmux 3.4",
	})
	ts := httptest.NewServer(srv.Router())
	return ts, memStore
}

// v1.REQ-005-AC-003：客户端连接 /ws 时接收初始 snapshot 快照
func TestWebSocket_SnapshotOnConnect(t *testing.T) {
	ts, memStore := setupTestWSServer("")
	defer ts.Close()

	// 预先写入一个 pane
	memStore.UpsertPane(model.PaneDetail{
		PaneSummary: model.PaneSummary{
			PaneID:       "%1",
			SessionName:  "dev",
			State:        model.StateWorking,
			LastActiveAt: time.Now(),
		},
	}, "init")

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("连接 WebSocket 失败: %v", err)
	}
	defer conn.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("期望状态码 101，实际为 %d", resp.StatusCode)
	}

	// 读取初始快照消息
	var msg struct {
		Type    string `json:"type"`
		Payload struct {
			Panes []model.PaneSummary `json:"panes"`
		} `json:"payload"`
	}
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("读取 snapshot 消息失败: %v", err)
	}

	if msg.Type != "snapshot" {
		t.Errorf("期望第一条消息为 snapshot，实际为 %s", msg.Type)
	}
	if len(msg.Payload.Panes) != 1 || msg.Payload.Panes[0].PaneID != "%1" {
		t.Errorf("快照数据不符: %+v", msg.Payload.Panes)
	}
}

// v1.REQ-005-AC-003：状态发生跃迁时广播推送 state_changed 事件
func TestWebSocket_StateChangedBroadcast(t *testing.T) {
	ts, memStore := setupTestWSServer("")
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("连接 WebSocket 失败: %v", err)
	}
	defer conn.Close()

	// 消费初始 snapshot
	var snapshot server.WSMessage
	_ = conn.ReadJSON(&snapshot)

	// 触发状态跃迁
	paneID := "%2"
	memStore.UpsertPane(model.PaneDetail{
		PaneSummary: model.PaneSummary{
			PaneID:       paneID,
			State:        model.StateWaitingInput,
			LastActiveAt: time.Now(),
		},
	}, "regex matched waiting_input")

	// 读取广播事件
	var broadcastMsg struct {
		Type    string `json:"type"`
		Payload struct {
			PaneID      string      `json:"pane_id"`
			ToState     model.State `json:"to_state"`
			TriggerRule string      `json:"trigger_rule"`
		} `json:"payload"`
	}
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if err := conn.ReadJSON(&broadcastMsg); err != nil {
		t.Fatalf("读取广播事件失败: %v", err)
	}

	if broadcastMsg.Type != "state_changed" {
		t.Errorf("期望事件类型为 state_changed，实际为 %s", broadcastMsg.Type)
	}
	if broadcastMsg.Payload.PaneID != paneID || broadcastMsg.Payload.ToState != model.StateWaitingInput {
		t.Errorf("广播载荷内容不符: %+v", broadcastMsg.Payload)
	}
}

// v1.REQ-005-AC-004：WebSocket 连接 Token 鉴权
func TestWebSocket_TokenAuth(t *testing.T) {
	const validToken = "test-secret-token-123456"
	ts, _ := setupTestWSServer(validToken)
	defer ts.Close()

	wsBase := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"

	t.Run("无 Token 连接失败返回 401", func(t *testing.T) {
		_, resp, err := websocket.DefaultDialer.Dial(wsBase, nil)
		if err == nil {
			t.Fatal("未携带 Token 应连接失败")
		}
		if resp != nil && resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("期望 401，实际为 %d", resp.StatusCode)
		}
	})

	t.Run("携带合法 Token 连接成功", func(t *testing.T) {
		conn, resp, err := websocket.DefaultDialer.Dial(wsBase+"?token="+validToken, nil)
		if err != nil {
			t.Fatalf("携带合法 Token 应当连接成功: %v", err)
		}
		defer conn.Close()
		if resp.StatusCode != http.StatusSwitchingProtocols {
			t.Errorf("期望 101，实际为 %d", resp.StatusCode)
		}
	})
}

// 客户端 Ping / Pong 心跳保活
func TestWebSocket_PingPong(t *testing.T) {
	ts, _ := setupTestWSServer("")
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("连接 WebSocket 失败: %v", err)
	}
	defer conn.Close()

	// 消费 snapshot
	var dump json.RawMessage
	_ = conn.ReadJSON(&dump)

	// 发送 ping action
	pingMsg := map[string]string{"action": "ping"}
	if err := conn.WriteJSON(pingMsg); err != nil {
		t.Fatalf("发送 ping 失败: %v", err)
	}

	// 接收 pong 响应
	var pongMsg struct {
		Type string `json:"type"`
	}
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if err := conn.ReadJSON(&pongMsg); err != nil {
		t.Fatalf("读取 pong 响应失败: %v", err)
	}
	if pongMsg.Type != "pong" {
		t.Errorf("期望回复 pong，实际为 %s", pongMsg.Type)
	}
}

