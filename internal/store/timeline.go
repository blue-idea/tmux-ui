package store

import (
	"tmux-ui/internal/model"
)

const MaxTimelineCapacity = 100

// TimelineRingBuffer 维护固定容量上限的时间线事件环形缓冲区
type TimelineRingBuffer struct {
	events []model.TimelineEvent
}

// NewTimelineRingBuffer 构造环形缓冲区
func NewTimelineRingBuffer() *TimelineRingBuffer {
	return &TimelineRingBuffer{
		events: make([]model.TimelineEvent, 0, MaxTimelineCapacity),
	}
}

// Append 添加一条时间线事件，超出容量自动丢弃最旧记录
func (b *TimelineRingBuffer) Append(evt model.TimelineEvent) {
	if len(b.events) < MaxTimelineCapacity {
		b.events = append(b.events, evt)
		return
	}
	// 滑动窗口丢弃头部
	b.events = append(b.events[1:], evt)
}

// Events 返回时间线事件的完整切片副本
func (b *TimelineRingBuffer) Events() []model.TimelineEvent {
	res := make([]model.TimelineEvent, len(b.events))
	copy(res, b.events)
	return res
}
