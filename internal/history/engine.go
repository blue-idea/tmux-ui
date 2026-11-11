package history

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"tmux-ui/internal/config"
	"tmux-ui/internal/model"
	"tmux-ui/internal/store"
)

// HistoryEngine 负责监听状态变迁并持久化历史快照，执行周期性 Checkpoint 保底
type HistoryEngine struct {
	cfg        config.HistoryConfig
	storage    *JSONLStorage
	store      *store.MemoryStore
	lastHashes map[string]string // 记录每个 Pane 最近一次保存快照时的输出哈希
	hashMu     sync.RWMutex
	stopCh     chan struct{}
}

// NewHistoryEngine 创建 HistoryEngine 实例
func NewHistoryEngine(cfg config.HistoryConfig, storage *JSONLStorage, memStore *store.MemoryStore) *HistoryEngine {
	if cfg.CheckpointIntervalS <= 0 {
		cfg.CheckpointIntervalS = 30
	}
	return &HistoryEngine{
		cfg:        cfg,
		storage:    storage,
		store:      memStore,
		lastHashes: make(map[string]string),
		stopCh:     make(chan struct{}),
	}
}

// Start 启动后台事件监听协程与周期性 Checkpoint 协程
func (e *HistoryEngine) Start(ctx context.Context) error {
	if !e.cfg.Enabled {
		return nil
	}

	eventCh := e.store.Subscribe(128)

	// 监听状态跃迁事件
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-e.stopCh:
				return
			case evt, ok := <-eventCh:
				if !ok {
					return
				}
				if evt.Type == store.EventStateChanged {
					e.handleStateChanged(evt)
				}
			}
		}
	}()

	// 周期性 Checkpoint 定时器
	go func() {
		ticker := time.NewTicker(time.Duration(e.cfg.CheckpointIntervalS) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-e.stopCh:
				return
			case <-ticker.C:
				e.TriggerCheckpoint()
			}
		}
	}()

	return nil
}

// Stop 停止引擎
func (e *HistoryEngine) Stop() {
	select {
	case <-e.stopCh:
	default:
		close(e.stopCh)
	}
}

// handleStateChanged 处理状态跃迁并记录快照
func (e *HistoryEngine) handleStateChanged(evt store.StoreEvent) {
	detail, exists := e.store.GetPane(evt.PaneID)
	if !exists {
		return
	}

	ts := detail.LastActiveAt.UnixMilli()
	if ts <= 0 {
		ts = time.Now().UnixMilli()
	}

	record := &model.SnapshotRecord{
		SnapshotMetadata: model.SnapshotMetadata{
			SnapshotID: fmt.Sprintf("snap_%s_%d", SanitizePaneID(detail.PaneID), ts),
			PaneID:     detail.PaneID,
			Timestamp:  ts,
			EventType:  "state_transition",
			OldState:   evt.FromState,
			NewState:   evt.ToState,
			Summary:    fmt.Sprintf("State transitioned to %s via %s", evt.ToState, evt.TriggerRule),
			ByteLength: len(detail.RawOutputANSI),
		},
		SessionName:  detail.SessionName,
		WindowID:     detail.WindowID,
		OutputBuffer: detail.RawOutputANSI,
		TailLines:    detail.PreviewLines,
	}

	_ = e.storage.AppendSnapshot(record)

	// 记录哈希
	h := computeOutputHash(detail.RawOutputANSI)
	e.hashMu.Lock()
	e.lastHashes[detail.PaneID] = h
	e.hashMu.Unlock()
}

// TriggerCheckpoint 检查当前活跃 Pane 的输出哈希变动，若发生变化则落盘 Checkpoint 快照
func (e *HistoryEngine) TriggerCheckpoint() int {
	panes := e.store.ListPanes()
	count := 0

	for _, summary := range panes {
		detail, exists := e.store.GetPane(summary.PaneID)
		if !exists || detail.IsDead || detail.State == model.StateExited {
			continue
		}

		currentHash := computeOutputHash(detail.RawOutputANSI)

		e.hashMu.RLock()
		lastHash, hasLast := e.lastHashes[detail.PaneID]
		e.hashMu.RUnlock()

		// 若已有记录且哈希未变动，则跳过
		if hasLast && currentHash == lastHash {
			continue
		}

		ts := detail.LastActiveAt.UnixMilli()
		if ts <= 0 {
			ts = time.Now().UnixMilli()
		}

		record := &model.SnapshotRecord{
			SnapshotMetadata: model.SnapshotMetadata{
				SnapshotID: fmt.Sprintf("snap_%s_%d_cp", SanitizePaneID(detail.PaneID), ts),
				PaneID:     detail.PaneID,
				Timestamp:  ts,
				EventType:  "checkpoint",
				OldState:   detail.State,
				NewState:   detail.State,
				Summary:    fmt.Sprintf("Periodic checkpoint in state %s", detail.State),
				ByteLength: len(detail.RawOutputANSI),
			},
			SessionName:  detail.SessionName,
			WindowID:     detail.WindowID,
			OutputBuffer: detail.RawOutputANSI,
			TailLines:    detail.PreviewLines,
		}

		if err := e.storage.AppendSnapshot(record); err == nil {
			e.hashMu.Lock()
			e.lastHashes[detail.PaneID] = currentHash
			e.hashMu.Unlock()
			count++
		}
	}

	return count
}

// computeOutputHash 计算终端输出的 SHA256 哈希值
func computeOutputHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}
