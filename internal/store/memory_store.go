package store

import (
	"fmt"
	"sync"
	"time"

	"tmux-ui/internal/model"
)

// EventType 表示内存状态变更事件类型
type EventType string

const (
	EventStateChanged EventType = "state_changed"
	EventPaneUpdated  EventType = "pane_updated"
)

// StoreEvent 广播的状态变更事件载荷
type StoreEvent struct {
	Type        EventType         `json:"type"`
	PaneID      string            `json:"pane_id"`
	FromState   model.State       `json:"from_state"`
	ToState     model.State       `json:"to_state"`
	TriggerRule string            `json:"trigger_rule"`
	Timestamp   time.Time         `json:"timestamp"`
	Summary     model.PaneSummary `json:"summary"`
}

// MemoryStore 内存状态中心，支持并发安全读写与事件订阅
type MemoryStore struct {
	mu          sync.RWMutex
	panes       map[string]*model.PaneDetail
	timelines   map[string]*TimelineRingBuffer
	subscribers []chan StoreEvent
}

// NewMemoryStore 构造 MemoryStore 实例
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		panes:     make(map[string]*model.PaneDetail),
		timelines: make(map[string]*TimelineRingBuffer),
	}
}

// UpsertPane 更新或插入 Pane 数据，若检测到状态迁移则记录 TimelineEvent 并广播事件
func (s *MemoryStore) UpsertPane(detail model.PaneDetail, triggerRule string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	paneID := detail.PaneID
	existing, exists := s.panes[paneID]
	timelineBuf, hasBuf := s.timelines[paneID]
	if !hasBuf {
		timelineBuf = NewTimelineRingBuffer()
		s.timelines[paneID] = timelineBuf
	}

	var fromState model.State
	stateChanged := false

	if !exists {
		fromState = detail.State
		stateChanged = true
	} else {
		fromState = existing.State
		if existing.State != detail.State {
			stateChanged = true
		}
	}

	// 发生状态变迁时追加时间线事件
	if stateChanged {
		evt := model.TimelineEvent{
			ID:          fmt.Sprintf("evt-%d", time.Now().UnixNano()),
			PaneID:      paneID,
			Timestamp:   detail.LastActiveAt,
			FromState:   fromState,
			ToState:     detail.State,
			TriggerRule: triggerRule,
		}
		timelineBuf.Append(evt)
	}

	// 挂载最新时间线切片副本
	detail.Timeline = timelineBuf.Events()
	s.panes[paneID] = &detail

	// 广播事件给订阅者
	if stateChanged {
		event := StoreEvent{
			Type:        EventStateChanged,
			PaneID:      paneID,
			FromState:   fromState,
			ToState:     detail.State,
			TriggerRule: triggerRule,
			Timestamp:   detail.LastActiveAt,
			Summary:     detail.PaneSummary,
		}
		s.broadcast(event)
	}
}

// GetPane 获取单个 Pane 完整详情
func (s *MemoryStore) GetPane(paneID string) (model.PaneDetail, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	detail, exists := s.panes[paneID]
	if !exists {
		return model.PaneDetail{}, false
	}
	// 返回副本
	cp := *detail
	if buf, ok := s.timelines[paneID]; ok {
		cp.Timeline = buf.Events()
	}
	return cp, true
}

// ListPanes 获取所有活动的 PaneSummary 列表
func (s *MemoryStore) ListPanes() []model.PaneSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]model.PaneSummary, 0, len(s.panes))
	for _, p := range s.panes {
		list = append(list, p.PaneSummary)
	}
	return list
}

// Subscribe 注册事件监听通道
func (s *MemoryStore) Subscribe(bufferSize int) <-chan StoreEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	if bufferSize <= 0 {
		bufferSize = 32
	}
	ch := make(chan StoreEvent, bufferSize)
	s.subscribers = append(s.subscribers, ch)
	return ch
}

// broadcast 内部广播事件给所有活动订阅者（非阻塞发送）
func (s *MemoryStore) broadcast(evt StoreEvent) {
	for _, ch := range s.subscribers {
		select {
		case ch <- evt:
		default:
			// 缓冲区满丢弃或慢消费者跳过
		}
	}
}

// CleanupDeadPanes 清理已处于 exited 状态且超过 TTL 的 Pane，返回清理的 PaneID 列表
func (s *MemoryStore) CleanupDeadPanes(ttl time.Duration, now time.Time) []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var removed []string
	for id, p := range s.panes {
		if p.IsDead || p.State == model.StateExited {
			if now.Sub(p.LastActiveAt) >= ttl {
				delete(s.panes, id)
				delete(s.timelines, id)
				removed = append(removed, id)
			}
		}
	}
	return removed
}
