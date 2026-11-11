package collector

import (
	"context"
	"fmt"
	"hash/fnv"
	"strings"
	"sync"
	"time"
)

// CapturePane 调用 tmux capture-pane -p -e -t %id 获取指定 pane 的完整 ANSI 输出
func (c *TmuxCollector) CapturePane(ctx context.Context, paneID string) (string, error) {
	args := c.BuildArgs("capture-pane", "-p", "-e", "-t", paneID)
	out, err := c.executor.Execute(ctx, args...)
	if err != nil {
		if isServerUnavailable(out, err) {
			return "", ErrTmuxUnavailable
		}
		return "", fmt.Errorf("tmux capture-pane failed: %w", err)
	}
	return string(out), nil
}

// ExtractTailLines 从原始终端输出中按行提取末尾 N 行文本（过滤空行）
func ExtractTailLines(text string, n int) []string {
	if text == "" || n <= 0 {
		return []string{}
	}

	// 统一换行符
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	rawLines := strings.Split(normalized, "\n")

	var validLines []string
	for _, l := range rawLines {
		trimmed := strings.TrimRight(l, " \t\r")
		if trimmed != "" {
			validLines = append(validLines, trimmed)
		}
	}

	total := len(validLines)
	if total <= n {
		return validLines
	}
	return validLines[total-n:]
}

// ComputeHash64 计算文本的 64 位 FNV-1a 哈希
func ComputeHash64(text string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(text))
	return h.Sum64()
}

// PaneOutputInfo 记录单个 Pane 的终端输出状态
type PaneOutputInfo struct {
	PaneID       string
	RawOutput    string
	TailLines    []string
	LastHash     uint64
	LastActiveAt time.Time
	IsFinalized  bool
}

// PaneOutputTracker 跟踪所有 Pane 的输出变化与活跃时间
type PaneOutputTracker struct {
	mu        sync.RWMutex
	tailLines int
	cache     map[string]*PaneOutputInfo
}

// NewPaneOutputTracker 构造输出追踪器
func NewPaneOutputTracker(tailLines int) *PaneOutputTracker {
	if tailLines <= 0 {
		tailLines = 50
	}
	return &PaneOutputTracker{
		tailLines: tailLines,
		cache:     make(map[string]*PaneOutputInfo),
	}
}

// Update 更新并研判 Pane 的最新输出
func (t *PaneOutputTracker) Update(paneID string, isDead bool, output string, now time.Time) (PaneOutputInfo, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	info, exists := t.cache[paneID]
	if exists && info.IsFinalized {
		// 已终结的 Pane 直接返回固化缓存，不产生变动通知
		return *info, false
	}

	tail := ExtractTailLines(output, t.tailLines)
	joinedTail := strings.Join(tail, "\n")
	newHash := ComputeHash64(joinedTail)

	changed := false
	if !exists {
		info = &PaneOutputInfo{
			PaneID:       paneID,
			RawOutput:    output,
			TailLines:    tail,
			LastHash:     newHash,
			LastActiveAt: now,
			IsFinalized:  isDead,
		}
		t.cache[paneID] = info
		changed = true
	} else {
		if newHash != info.LastHash {
			info.LastHash = newHash
			info.LastActiveAt = now
			changed = true
		}
		info.RawOutput = output
		info.TailLines = tail
		if isDead {
			info.IsFinalized = true
		}
	}

	return *info, changed
}
