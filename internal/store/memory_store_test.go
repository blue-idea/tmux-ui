package store_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"tmux-ui/internal/model"
	"tmux-ui/internal/store"
)

// v1.REQ-003-AC-005：状态发生迁移时自动向该 Pane 追加 TimelineEvent 记录
func TestStateStore_TimelineAppendOnStateTransition(t *testing.T) {
	s := store.NewMemoryStore()
	paneID := "%1"
	t0 := time.Date(2026, 11, 11, 20, 0, 0, 0, time.UTC)

	// 初始状态写入
	p1 := model.PaneDetail{
		PaneSummary: model.PaneSummary{
			PaneID:       paneID,
			State:        model.StateIdle,
			LastActiveAt: t0,
		},
	}
	s.UpsertPane(p1, "initial state")

	detail, exists := s.GetPane(paneID)
	if !exists {
		t.Fatal("未找到写入的 Pane")
	}
	if len(detail.Timeline) != 1 {
		t.Fatalf("初始写入应生成 1 条时间线记录，实际为 %d", len(detail.Timeline))
	}
	if detail.Timeline[0].ToState != model.StateIdle {
		t.Errorf("初始时间线状态错误: %s", detail.Timeline[0].ToState)
	}

	// 状态迁移到 working
	t1 := t0.Add(5 * time.Second)
	p2 := model.PaneDetail{
		PaneSummary: model.PaneSummary{
			PaneID:       paneID,
			State:        model.StateWorking,
			LastActiveAt: t1,
		},
	}
	s.UpsertPane(p2, "regex matched working")

	detail2, _ := s.GetPane(paneID)
	if len(detail2.Timeline) != 2 {
		t.Fatalf("状态迁移后应有 2 条时间线记录，实际为 %d", len(detail2.Timeline))
	}
	lastEvt := detail2.Timeline[1]
	if lastEvt.FromState != model.StateIdle || lastEvt.ToState != model.StateWorking {
		t.Errorf("时间线迁移状态错误: %+v", lastEvt)
	}
	if lastEvt.TriggerRule != "regex matched working" {
		t.Errorf("触发原因记录错误: %s", lastEvt.TriggerRule)
	}
}

// 环形缓冲容量测试：时间线记录上限 100 条
func TestStateStore_TimelineRingBufferBounded(t *testing.T) {
	s := store.NewMemoryStore()
	paneID := "%2"
	now := time.Now()

	// 写入 105 次不同的状态跃迁
	for i := 0; i < 105; i++ {
		st := model.StateWorking
		if i%2 == 0 {
			st = model.StateWaitingInput
		}
		p := model.PaneDetail{
			PaneSummary: model.PaneSummary{
				PaneID:       paneID,
				State:        st,
				LastActiveAt: now.Add(time.Duration(i) * time.Second),
			},
		}
		s.UpsertPane(p, fmt.Sprintf("transition-%d", i))
	}

	detail, _ := s.GetPane(paneID)
	if len(detail.Timeline) != 100 {
		t.Fatalf("时间线长度应严格限制为 100，实际为 %d", len(detail.Timeline))
	}

	// 验证时间线正序保持最新记录（最后一条应当是 transition-104）
	lastEvt := detail.Timeline[len(detail.Timeline)-1]
	if lastEvt.TriggerRule != "transition-104" {
		t.Errorf("最新时间线记录应为 transition-104，实际为 %s", lastEvt.TriggerRule)
	}
}

// 事件广播测试：状态变迁时触发事件通知
func TestStateStore_EventBroadcasting(t *testing.T) {
	s := store.NewMemoryStore()
	eventsChan := s.Subscribe(10)

	paneID := "%3"
	now := time.Now()

	p := model.PaneDetail{
		PaneSummary: model.PaneSummary{
			PaneID:       paneID,
			State:        model.StateWorking,
			LastActiveAt: now,
		},
	}
	s.UpsertPane(p, "started")

	select {
	case evt := <-eventsChan:
		if evt.PaneID != paneID || evt.Type != store.EventStateChanged {
			t.Errorf("接收到的事件不符: %+v", evt)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("超时未收到状态变迁事件")
	}
}

// 僵尸 Pane 清理：已退出超过 TTL 的 Pane 自动清理
func TestStateStore_CleanupDeadPanes(t *testing.T) {
	s := store.NewMemoryStore()
	now := time.Now()

	// Pane 1: 存活
	s.UpsertPane(model.PaneDetail{
		PaneSummary: model.PaneSummary{
			PaneID:       "%1",
			State:        model.StateWorking,
			LastActiveAt: now,
			IsDead:       false,
		},
	}, "alive")

	// Pane 2: 死亡但未超过 10 分钟 TTL (5 分钟前)
	s.UpsertPane(model.PaneDetail{
		PaneSummary: model.PaneSummary{
			PaneID:       "%2",
			State:        model.StateExited,
			LastActiveAt: now.Add(-5 * time.Minute),
			IsDead:       true,
		},
	}, "recently dead")

	// Pane 3: 死亡且超过 10 分钟 TTL (15 分钟前)
	s.UpsertPane(model.PaneDetail{
		PaneSummary: model.PaneSummary{
			PaneID:       "%3",
			State:        model.StateExited,
			LastActiveAt: now.Add(-15 * time.Minute),
			IsDead:       true,
		},
	}, "old dead")

	// 执行清理
	removed := s.CleanupDeadPanes(10*time.Minute, now)
	if len(removed) != 1 || removed[0] != "%3" {
		t.Fatalf("应只清理超过 10 分钟的 Pane 3，实际清理: %+v", removed)
	}

	if _, exists := s.GetPane("%3"); exists {
		t.Error("Pane 3 应该已被清理")
	}
	if _, exists := s.GetPane("%2"); !exists {
		t.Error("Pane 2 不应被清理")
	}
	if _, exists := s.GetPane("%1"); !exists {
		t.Error("Pane 1 不应被清理")
	}
}

// 并发读写安全测试
func TestStateStore_ConcurrentSafety(t *testing.T) {
	s := store.NewMemoryStore()
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			paneID := fmt.Sprintf("%%%d", id%5)
			for j := 0; j < 50; j++ {
				s.UpsertPane(model.PaneDetail{
					PaneSummary: model.PaneSummary{
						PaneID:       paneID,
						State:        model.StateWorking,
						LastActiveAt: time.Now(),
					},
				}, "concurrent")
				_ = s.ListPanes()
				_, _ = s.GetPane(paneID)
			}
		}(i)
	}

	wg.Wait()
}
