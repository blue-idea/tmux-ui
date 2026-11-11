package history_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"tmux-ui/internal/config"
	"tmux-ui/internal/history"
	"tmux-ui/internal/model"
)

// TestStorage_SanitizePaneID 验证 Pane ID 路径安全转义防路径穿越
func TestStorage_SanitizePaneID(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"%0", "_0"},
		{"%15", "_15"},
		{"../etc/passwd", "___etc_passwd"},
		{"pane/with/slash", "pane_with_slash"},
		{"normal-id", "normal-id"},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			got := history.SanitizePaneID(c.input)
			if got != c.expected {
				t.Fatalf("expected sanitized pane ID %q, got %q", c.expected, got)
			}
		})
	}
}

// TestStorage_AppendAndReadRecord 验证 SnapshotRecord 的追加写入与完整还原
func TestStorage_AppendAndReadRecord(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.HistoryConfig{
		Enabled:             true,
		StorageDir:          tempDir,
		MaxSnapshotsPerPane: 50,
		MaxBytesPerPane:     1024 * 1024,
	}

	storage, err := history.NewJSONLStorage(cfg)
	if err != nil {
		t.Fatalf("failed to create JSONLStorage: %v", err)
	}

	record := &model.SnapshotRecord{
		SnapshotMetadata: model.SnapshotMetadata{
			SnapshotID: "snap__0_1731380400000_1234",
			PaneID:     "%0",
			Timestamp:  1731380400000,
			EventType:  "state_transition",
			OldState:   model.StateWorking,
			NewState:   model.StateWaitingInput,
			Summary:    "Waiting for user confirmation",
			ByteLength: 35,
		},
		SessionName:  "dev",
		WindowID:     "@0",
		OutputBuffer: "\x1b[33mDo you want to proceed? (y/n)\x1b[0m",
		TailLines:    []string{"Do you want to proceed? (y/n)"},
	}

	// 写入快照
	if err := storage.AppendSnapshot(record); err != nil {
		t.Fatalf("AppendSnapshot failed: %v", err)
	}

	// 读取该快照
	readRecord, err := storage.GetSnapshot("%0", record.SnapshotID)
	if err != nil {
		t.Fatalf("GetSnapshot failed: %v", err)
	}

	if readRecord.SnapshotID != record.SnapshotID {
		t.Errorf("expected SnapshotID %s, got %s", record.SnapshotID, readRecord.SnapshotID)
	}
	if readRecord.OutputBuffer != record.OutputBuffer {
		t.Errorf("expected OutputBuffer %q, got %q", record.OutputBuffer, readRecord.OutputBuffer)
	}
	if readRecord.NewState != model.StateWaitingInput {
		t.Errorf("expected NewState %s, got %s", model.StateWaitingInput, readRecord.NewState)
	}
}

// TestStorage_QueryMetadataList 验证轻量级快照元数据列表查询
func TestStorage_QueryMetadataList(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.HistoryConfig{
		Enabled:             true,
		StorageDir:          tempDir,
		MaxSnapshotsPerPane: 100,
		MaxBytesPerPane:     1024 * 1024,
	}

	storage, err := history.NewJSONLStorage(cfg)
	if err != nil {
		t.Fatalf("failed to create JSONLStorage: %v", err)
	}

	baseTime := time.Now().UnixMilli()
	for i := 1; i <= 5; i++ {
		rec := &model.SnapshotRecord{
			SnapshotMetadata: model.SnapshotMetadata{
				SnapshotID: fmt.Sprintf("snap__0_%d", baseTime+int64(i*1000)),
				PaneID:     "%0",
				Timestamp:  baseTime + int64(i*1000),
				EventType:  "checkpoint",
				OldState:   model.StateWorking,
				NewState:   model.StateWorking,
				Summary:    fmt.Sprintf("Checkpoint %d", i),
				ByteLength: 20,
			},
			SessionName:  "main",
			WindowID:     "@0",
			OutputBuffer: fmt.Sprintf("Output line %d", i),
		}
		if err := storage.AppendSnapshot(rec); err != nil {
			t.Fatalf("failed to append snapshot %d: %v", i, err)
		}
	}

	// 查询全部列表
	list, err := storage.ListSnapshots("%0", 0, 10)
	if err != nil {
		t.Fatalf("ListSnapshots failed: %v", err)
	}
	if len(list) != 5 {
		t.Fatalf("expected 5 snapshots, got %d", len(list))
	}

	// 验证时间戳筛选 (since)
	sinceTime := baseTime + 2500
	filteredList, err := storage.ListSnapshots("%0", sinceTime, 10)
	if err != nil {
		t.Fatalf("ListSnapshots with since failed: %v", err)
	}
	if len(filteredList) != 3 {
		t.Fatalf("expected 3 snapshots after since %d, got %d", sinceTime, len(filteredList))
	}
}

// TestStorage_FIFOPruning 验证超出上限时的自动 FIFO 淘汰机制
func TestStorage_FIFOPruning(t *testing.T) {
	tempDir := t.TempDir()
	maxSnapshots := 5
	cfg := config.HistoryConfig{
		Enabled:             true,
		StorageDir:          tempDir,
		MaxSnapshotsPerPane: maxSnapshots,
		MaxBytesPerPane:     1024 * 1024,
	}

	storage, err := history.NewJSONLStorage(cfg)
	if err != nil {
		t.Fatalf("failed to create JSONLStorage: %v", err)
	}

	baseTime := time.Now().UnixMilli()
	totalToInsert := 12
	for i := 1; i <= totalToInsert; i++ {
		rec := &model.SnapshotRecord{
			SnapshotMetadata: model.SnapshotMetadata{
				SnapshotID: fmt.Sprintf("snap__1_%02d", i),
				PaneID:     "%1",
				Timestamp:  baseTime + int64(i*1000),
				EventType:  "state_transition",
				OldState:   model.StateWorking,
				NewState:   model.StateWorking,
				Summary:    fmt.Sprintf("Event %d", i),
				ByteLength: 15,
			},
			SessionName:  "dev",
			WindowID:     "@1",
			OutputBuffer: fmt.Sprintf("Buffer content %d", i),
		}
		if err := storage.AppendSnapshot(rec); err != nil {
			t.Fatalf("failed to insert record %d: %v", i, err)
		}
	}

	// 验证保留的记录数已被限制在 maxSnapshots
	list, err := storage.ListSnapshots("%1", 0, 100)
	if err != nil {
		t.Fatalf("ListSnapshots failed: %v", err)
	}
	if len(list) != maxSnapshots {
		t.Fatalf("expected %d snapshots after pruning, got %d", maxSnapshots, len(list))
	}

	// 确认保留的是最新的记录（8, 9, 10, 11, 12）
	expectedFirstID := fmt.Sprintf("snap__1_%02d", totalToInsert-maxSnapshots+1)
	if list[0].SnapshotID != expectedFirstID {
		t.Errorf("expected first retained snapshot to be %s, got %s", expectedFirstID, list[0].SnapshotID)
	}
}

// TestStorage_ConcurrentSafety 验证多协程并发写入安全无冲突与死锁
func TestStorage_ConcurrentSafety(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.HistoryConfig{
		Enabled:             true,
		StorageDir:          tempDir,
		MaxSnapshotsPerPane: 200,
		MaxBytesPerPane:     1024 * 1024,
	}

	storage, err := history.NewJSONLStorage(cfg)
	if err != nil {
		t.Fatalf("failed to create JSONLStorage: %v", err)
	}

	const goroutines = 10
	const perGoroutine = 10
	var wg sync.WaitGroup

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				paneID := fmt.Sprintf("%%concurrent_%d", gid%2) // 集中写入 2 个 pane 测试锁
				rec := &model.SnapshotRecord{
					SnapshotMetadata: model.SnapshotMetadata{
						SnapshotID: fmt.Sprintf("snap_c_%d_%d", gid, i),
						PaneID:     paneID,
						Timestamp:  time.Now().UnixMilli(),
						EventType:  "concurrent_test",
						NewState:   model.StateWorking,
					},
					OutputBuffer: "concurrent test payload",
				}
				if err := storage.AppendSnapshot(rec); err != nil {
					t.Errorf("concurrent append failed: %v", err)
				}
			}
		}(g)
	}

	wg.Wait()

	// 验证各 Pane 文件有效且可读取
	for _, paneID := range []string{"%concurrent_0", "%concurrent_1"} {
		list, err := storage.ListSnapshots(paneID, 0, 500)
		if err != nil {
			t.Fatalf("failed to list snapshots for %s: %v", paneID, err)
		}
		if len(list) == 0 {
			t.Errorf("expected snapshots for %s, got 0", paneID)
		}
	}
}


// TestStorage_EdgeCasesAndErrors 覆盖边界条件与异常路径
func TestStorage_EdgeCasesAndErrors(t *testing.T) {
	tempDir := t.TempDir()

	// 1. 默认配置自动填充
	storage, err := history.NewJSONLStorage(config.HistoryConfig{
		StorageDir: tempDir,
	})
	if err != nil {
		t.Fatalf("failed to init storage: %v", err)
	}

	// 2. 空记录或空 PaneID 拒绝写入
	if err := storage.AppendSnapshot(nil); err == nil {
		t.Errorf("expected error when appending nil record, got nil")
	}
	if err := storage.AppendSnapshot(&model.SnapshotRecord{}); err == nil {
		t.Errorf("expected error when appending record with empty PaneID, got nil")
	}

	// 3. 查询不存在的 Pane 返回空列表与 ErrSnapshotNotFound
	list, err := storage.ListSnapshots("%nonexistent", 0, 10)
	if err != nil {
		t.Fatalf("expected nil error for nonexistent pane list, got %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 snapshots for nonexistent pane, got %d", len(list))
	}

	_, err = storage.GetSnapshot("%nonexistent", "snap_dummy")
	if err != history.ErrSnapshotNotFound {
		t.Errorf("expected ErrSnapshotNotFound, got %v", err)
	}

	// 4. 写入一条记录后，查询该 Pane 下不存在的快照 ID
	rec := &model.SnapshotRecord{
		SnapshotMetadata: model.SnapshotMetadata{
			PaneID: "%existing",
		},
		OutputBuffer: "some output",
	}
	if err := storage.AppendSnapshot(rec); err != nil {
		t.Fatalf("failed to append snapshot: %v", err)
	}

	_, err = storage.GetSnapshot("%existing", "snap_not_found")
	if err != history.ErrSnapshotNotFound {
		t.Errorf("expected ErrSnapshotNotFound, got %v", err)
	}
}
