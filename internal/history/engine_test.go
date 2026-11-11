package history_test

import (
	"context"
	"testing"
	"time"

	"tmux-ui/internal/config"
	"tmux-ui/internal/history"
	"tmux-ui/internal/model"
	"tmux-ui/internal/store"
)

// TestEngine_StateTransitionCapturesSnapshot 验证状态变迁事件触发快照自动捕获与持久化
func TestEngine_StateTransitionCapturesSnapshot(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.HistoryConfig{
		Enabled:             true,
		StorageDir:          tempDir,
		MaxSnapshotsPerPane: 50,
		MaxBytesPerPane:     1024 * 1024,
		CheckpointIntervalS: 10,
	}

	storage, err := history.NewJSONLStorage(cfg)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	memStore := store.NewMemoryStore()
	engine := history.NewHistoryEngine(cfg, storage, memStore)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := engine.Start(ctx); err != nil {
		t.Fatalf("failed to start history engine: %v", err)
	}

	// 模拟初始 Pane
	now := time.Now()
	detail := model.PaneDetail{
		PaneSummary: model.PaneSummary{
			PaneID:       "%0",
			SessionName:  "dev",
			WindowID:     "@0",
			State:        model.StateWorking,
			LastActiveAt: now,
		},
		RawOutputANSI: "Initial working output",
	}
	detail.PreviewLines = []string{"Initial working output"}
	memStore.UpsertPane(detail, "test_rule")

	// 发生状态变迁至 waiting_input
	detail.State = model.StateWaitingInput
	detail.RawOutputANSI = "Do you want to proceed? (y/n)"
	detail.PreviewLines = []string{"Do you want to proceed? (y/n)"}
	detail.LastActiveAt = now.Add(time.Second)
	memStore.UpsertPane(detail, "confirm_rule")

	// 等待异步处理
	time.Sleep(100 * time.Millisecond)

	// 验证存储中存在快照
	snapshots, err := storage.ListSnapshots("%0", 0, 10)
	if err != nil {
		t.Fatalf("failed to list snapshots: %v", err)
	}

	if len(snapshots) == 0 {
		t.Fatalf("expected at least 1 snapshot after state transition, got 0")
	}

	lastSnap := snapshots[len(snapshots)-1]
	if lastSnap.NewState != model.StateWaitingInput {
		t.Errorf("expected snapshot state %s, got %s", model.StateWaitingInput, lastSnap.NewState)
	}
	if lastSnap.EventType != "state_transition" {
		t.Errorf("expected EventType 'state_transition', got %s", lastSnap.EventType)
	}

	record, err := storage.GetSnapshot("%0", lastSnap.SnapshotID)
	if err != nil {
		t.Fatalf("failed to retrieve snapshot record: %v", err)
	}
	if record.OutputBuffer != detail.RawOutputANSI {
		t.Errorf("expected OutputBuffer %q, got %q", detail.RawOutputANSI, record.OutputBuffer)
	}
}

// TestEngine_PeriodicCheckpoint 验证输出发生变动时保底 Checkpoint 自动触发
func TestEngine_PeriodicCheckpoint(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.HistoryConfig{
		Enabled:             true,
		StorageDir:          tempDir,
		MaxSnapshotsPerPane: 50,
		MaxBytesPerPane:     1024 * 1024,
		CheckpointIntervalS: 1, // 1秒短周期便于单元测试
	}

	storage, err := history.NewJSONLStorage(cfg)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	memStore := store.NewMemoryStore()
	engine := history.NewHistoryEngine(cfg, storage, memStore)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := engine.Start(ctx); err != nil {
		t.Fatalf("failed to start history engine: %v", err)
	}

	now := time.Now()
	detail := model.PaneDetail{
		PaneSummary: model.PaneSummary{
			PaneID:       "%1",
			SessionName:  "prod",
			WindowID:     "@1",
			State:        model.StateWorking,
			LastActiveAt: now,
		},
		RawOutputANSI: "Processing step 1...",
	}
	detail.PreviewLines = []string{"Processing step 1..."}
	memStore.UpsertPane(detail, "init")

	// 消费初始变迁快照
	time.Sleep(100 * time.Millisecond)

	// 更新输出但保持 working 状态（未触发状态变迁）
	detail.RawOutputANSI = "Processing step 2...\nProcessing step 3..."
	detail.PreviewLines = []string{"Processing step 3..."}
	detail.LastActiveAt = now.Add(2 * time.Second)
	memStore.UpsertPane(detail, "progress")

	// 手动触发一次 Checkpoint 检查
	checkpointCount := engine.TriggerCheckpoint()
	if checkpointCount < 1 {
		t.Fatalf("expected at least 1 checkpoint triggered, got %d", checkpointCount)
	}

	// 检查存储中是否有 checkpoint 快照
	snapshots, err := storage.ListSnapshots("%1", 0, 10)
	if err != nil {
		t.Fatalf("failed to list snapshots: %v", err)
	}

	foundCheckpoint := false
	for _, s := range snapshots {
		if s.EventType == "checkpoint" {
			foundCheckpoint = true
			break
		}
	}
	if !foundCheckpoint {
		t.Errorf("expected to find snapshot with event_type 'checkpoint'")
	}
}

// TestEngine_DeadPanePreserved 验证销毁的 Pane 历史快照安全留存不丢失
func TestEngine_DeadPanePreserved(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.HistoryConfig{
		Enabled:             true,
		StorageDir:          tempDir,
		MaxSnapshotsPerPane: 50,
		MaxBytesPerPane:     1024 * 1024,
		CheckpointIntervalS: 30,
	}

	storage, err := history.NewJSONLStorage(cfg)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	memStore := store.NewMemoryStore()
	engine := history.NewHistoryEngine(cfg, storage, memStore)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("failed to start engine: %v", err)
	}

	detail := model.PaneDetail{
		PaneSummary: model.PaneSummary{
			PaneID:       "%2",
			SessionName:  "temp",
			WindowID:     "@2",
			State:        model.StateWorking,
			LastActiveAt: time.Now(),
		},
		RawOutputANSI: "Task finished",
	}
	memStore.UpsertPane(detail, "start")

	detail.State = model.StateExited
	detail.IsDead = true
	memStore.UpsertPane(detail, "pane_dead")

	time.Sleep(100 * time.Millisecond)

	// 模拟内存清理已销毁的 Pane
	memStore.CleanupDeadPanes(0, time.Now().Add(time.Hour))

	// 验证内存中 Pane 已移除
	if _, exists := memStore.GetPane("%2"); exists {
		t.Fatalf("expected pane to be removed from memory store")
	}

	// 验证磁盘中的快照依然完整可查
	snapshots, err := storage.ListSnapshots("%2", 0, 10)
	if err != nil {
		t.Fatalf("failed to list snapshots after cleanup: %v", err)
	}
	if len(snapshots) == 0 {
		t.Fatalf("expected snapshots to persist on disk after memory cleanup, got 0")
	}
}
