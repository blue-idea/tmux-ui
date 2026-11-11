package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tmux-ui/internal/config"
	"tmux-ui/internal/history"
	"tmux-ui/internal/model"
	"tmux-ui/internal/server"
	"tmux-ui/internal/store"
)

// setupHistoryTestServer 构造带有历史存储支持的测试 HTTP 服务器
func setupHistoryTestServer(t *testing.T, token string) (*server.HTTPServer, *history.JSONLStorage, *store.MemoryStore) {
	tempDir := t.TempDir()
	cfg := config.HistoryConfig{
		Enabled:             true,
		StorageDir:          tempDir,
		MaxSnapshotsPerPane: 50,
		MaxBytesPerPane:     1024 * 1024,
	}

	storage, err := history.NewJSONLStorage(cfg)
	if err != nil {
		t.Fatalf("failed to init storage: %v", err)
	}

	memStore := store.NewMemoryStore()
	hub := server.NewWSHub(memStore, token)
	go hub.Run()
	t.Cleanup(func() { hub.Close() })

	srv := server.NewHTTPServer(server.HTTPServerOptions{
		Store:          memStore,
		Config:         config.DefaultConfig(),
		AuthToken:      token,
		WSHub:          hub,
		HistoryStorage: storage,
	})

	return srv, storage, memStore
}

// TestSnapshotAPI_ListSnapshots 验证拉取快照元数据列表接口
func TestSnapshotAPI_ListSnapshots(t *testing.T) {
	srv, storage, _ := setupHistoryTestServer(t, "")

	// 插入两条测试快照
	rec1 := &model.SnapshotRecord{
		SnapshotMetadata: model.SnapshotMetadata{
			SnapshotID: "snap_0_1001",
			PaneID:     "%0",
			Timestamp:  1000,
			EventType:  "state_transition",
			OldState:   model.StateIdle,
			NewState:   model.StateWorking,
			Summary:    "Started",
			ByteLength: 10,
		},
		OutputBuffer: "Output 1",
	}
	rec2 := &model.SnapshotRecord{
		SnapshotMetadata: model.SnapshotMetadata{
			SnapshotID: "snap_0_1002",
			PaneID:     "%0",
			Timestamp:  2000,
			EventType:  "checkpoint",
			OldState:   model.StateWorking,
			NewState:   model.StateWorking,
			Summary:    "Checkpoint",
			ByteLength: 20,
		},
		OutputBuffer: "Output 2",
	}

	_ = storage.AppendSnapshot(rec1)
	_ = storage.AppendSnapshot(rec2)

	// 发起 GET /api/v1/panes/%0/snapshots
	req := httptest.NewRequest(http.MethodGet, "/api/v1/panes/%250/snapshots", nil)
	rec := httptest.NewRecorder()
	srv.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Code    int                      `json:"code"`
		Message string                   `json:"message"`
		Data    []model.SnapshotMetadata `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if resp.Code != 0 || len(resp.Data) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(resp.Data))
	}
	if resp.Data[0].SnapshotID != "snap_0_1001" || resp.Data[1].SnapshotID != "snap_0_1002" {
		t.Errorf("unexpected snapshot order or IDs: %+v", resp.Data)
	}

	// 验证裸数字路由兼容 /api/v1/panes/0/snapshots
	reqRaw := httptest.NewRequest(http.MethodGet, "/api/v1/panes/0/snapshots", nil)
	recRaw := httptest.NewRecorder()
	srv.Router().ServeHTTP(recRaw, reqRaw)

	if recRaw.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 for raw ID /panes/0/snapshots, got %d", recRaw.Code)
	}
}

// TestSnapshotAPI_GetSnapshotDetail 验证获取单条快照与 404 错误
func TestSnapshotAPI_GetSnapshotDetail(t *testing.T) {
	srv, storage, _ := setupHistoryTestServer(t, "")

	rec := &model.SnapshotRecord{
		SnapshotMetadata: model.SnapshotMetadata{
			SnapshotID: "snap_1_target",
			PaneID:     "%1",
			Timestamp:  time.Now().UnixMilli(),
			EventType:  "state_transition",
			OldState:   model.StateWorking,
			NewState:   model.StateWaitingInput,
			Summary:    "Need input",
			ByteLength: 25,
		},
		OutputBuffer: "\x1b[33mConfirm? (y/n)\x1b[0m",
		TailLines:    []string{"Confirm? (y/n)"},
	}
	_ = storage.AppendSnapshot(rec)

	// 1. 查询存在的快照
	req := httptest.NewRequest(http.MethodGet, "/api/v1/panes/%251/snapshots/snap_1_target", nil)
	respRec := httptest.NewRecorder()
	srv.Router().ServeHTTP(respRec, req)

	if respRec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d: %s", respRec.Code, respRec.Body.String())
	}

	var detailResp struct {
		Code int                  `json:"code"`
		Data model.SnapshotRecord `json:"data"`
	}
	if err := json.Unmarshal(respRec.Body.Bytes(), &detailResp); err != nil {
		t.Fatalf("failed to decode JSON detail: %v", err)
	}
	if detailResp.Data.OutputBuffer != rec.OutputBuffer {
		t.Errorf("expected OutputBuffer %q, got %q", rec.OutputBuffer, detailResp.Data.OutputBuffer)
	}

	// 2. 查询不存在的快照返回 404 及标准错误信封
	req404 := httptest.NewRequest(http.MethodGet, "/api/v1/panes/%251/snapshots/nonexistent_snap", nil)
	resp404 := httptest.NewRecorder()
	srv.Router().ServeHTTP(resp404, req404)

	if resp404.Code != http.StatusNotFound {
		t.Fatalf("expected HTTP 404, got %d", resp404.Code)
	}

	var errEnvelope struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(resp404.Body.Bytes(), &errEnvelope); err != nil {
		t.Fatalf("failed to unmarshal error envelope: %v", err)
	}
	if errEnvelope.Error.Code != "SNAPSHOT_NOT_FOUND" {
		t.Errorf("expected error code 'SNAPSHOT_NOT_FOUND', got %s", errEnvelope.Error.Code)
	}
}

// TestSnapshotAPI_WSHubBroadcast 验证 WebSocket 广播 snapshot_created 帧
func TestSnapshotAPI_WSHubBroadcast(t *testing.T) {
	memStore := store.NewMemoryStore()
	hub := server.NewWSHub(memStore, "")
	go hub.Run()
	defer hub.Close()

	// 广播一条 snapshot_created
	meta := model.SnapshotMetadata{
		SnapshotID: "snap_broadcast",
		PaneID:     "%0",
		Timestamp:  12345,
		EventType:  "state_transition",
	}
	hub.BroadcastMessage("snapshot_created", meta)

	// 验证未崩溃且能正常处理
	time.Sleep(50 * time.Millisecond)
}
