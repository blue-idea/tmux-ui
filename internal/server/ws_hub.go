package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"tmux-ui/internal/store"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 本地工具允许跨源连接
	},
}

// WSMessage 统一推流消息帧格式
type WSMessage struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

// WSClient 封装单个 WebSocket 客户端连接
type WSClient struct {
	hub  *WSHub
	conn *websocket.Conn
	send chan []byte
}

// WSHub 管理所有客户端连接与全局事件广播
type WSHub struct {
	store       *store.MemoryStore
	authToken   string
	clients     map[*WSClient]bool
	broadcast   chan WSMessage
	register    chan *WSClient
	unregister  chan *WSClient
	mu          sync.RWMutex
	stopChan    chan struct{}
}

// NewWSHub 构造 WSHub 实例
func NewWSHub(memStore *store.MemoryStore, authToken string) *WSHub {
	hub := &WSHub{
		store:      memStore,
		authToken:  authToken,
		clients:    make(map[*WSClient]bool),
		broadcast:  make(chan WSMessage, 64),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
		stopChan:   make(chan struct{}),
	}
	return hub
}

// Run 启动 WebSocket Hub 事件循环
func (h *WSHub) Run() {
	var storeEvents <-chan store.StoreEvent
	if h.store != nil {
		storeEvents = h.store.Subscribe(64)
	}

	for {
		select {
		case <-h.stopChan:
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.broadcastBytes(msg)

		case storeEvt, ok := <-storeEvents:
			if !ok {
				continue
			}
			// 接收 StateStore 状态迁移，广播给客户端
			msg := WSMessage{
				Type:      string(storeEvt.Type),
				Timestamp: storeEvt.Timestamp,
				Payload: map[string]interface{}{
					"pane_id":      storeEvt.PaneID,
					"from_state":   storeEvt.FromState,
					"to_state":     storeEvt.ToState,
					"trigger_rule": storeEvt.TriggerRule,
					"summary":      storeEvt.Summary,
				},
			}
			h.broadcastBytes(msg)
		}
	}
}

// Close 关闭 Hub
func (h *WSHub) Close() {
	close(h.stopChan)
}

// BroadcastMessage 向所有连接的 WebSocket 客户端广播自定义消息（如 snapshot_created）
func (h *WSHub) BroadcastMessage(msgType string, payload interface{}) {
	select {
	case h.broadcast <- WSMessage{
		Type:      msgType,
		Timestamp: time.Now(),
		Payload:   payload,
	}:
	default:
	}
}

// broadcastBytes 将消息序列化为 JSON 并分发给所有客户端
func (h *WSHub) broadcastBytes(msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.send <- data:
		default:
			// 慢消费者非阻塞处理
		}
	}
}

// HandleWS 处理 WebSocket 升级与握手
func (h *WSHub) HandleWS(w http.ResponseWriter, r *http.Request) {
	// Token 鉴权校验
	if h.authToken != "" {
		token := r.URL.Query().Get("token")
		if token == "" {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}
		if token != h.authToken {
			renderError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication token is missing or invalid", nil)
			return
		}
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("WebSocket upgrade failed", "error", err)
		return
	}

	client := &WSClient{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 64),
	}
	h.register <- client

	// 连入后立即发送全量 snapshot 快照
	panes := []interface{}{}
	if h.store != nil {
		for _, p := range h.store.ListPanes() {
			panes = append(panes, p)
		}
	}
	snapshotMsg := WSMessage{
		Type:      "snapshot",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"panes": panes,
		},
	}
	if snapBytes, err := json.Marshal(snapshotMsg); err == nil {
		client.send <- snapBytes
	}

	go client.writePump()
	go client.readPump()
}

// readPump 监听客户端上行消息（心跳 Ping 等）
func (c *WSClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var payload struct {
			Action string `json:"action"`
		}
		if err := json.Unmarshal(message, &payload); err == nil {
			if payload.Action == "ping" {
				pong := WSMessage{
					Type:      "pong",
					Timestamp: time.Now(),
				}
				if pongBytes, err := json.Marshal(pong); err == nil {
					c.send <- pongBytes
				}
			}
		}
	}
}

// writePump 负责向 WebSocket 连接下发数据
func (c *WSClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

