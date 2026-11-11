package collector_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"tmux-ui/internal/collector"
)

// v1.REQ-002-AC-001：执行 capture-pane -p -e -t %id 并保留 ANSI 彩色转义码
func TestCapturePane_ANSIEscapeRetained(t *testing.T) {
	ansiOutput := "\x1b[32m[Agent]\x1b[0m Thinking...\n\x1b[33mDo you want to proceed? (y/n)\x1b[0m\n"
	mock := &MockTmuxExecutor{
		Outputs: map[string][]byte{
			"capture-pane": []byte(ansiOutput),
		},
	}

	c := collector.NewTmuxCollector(collector.Options{Executor: mock})
	out, err := c.CapturePane(context.Background(), "%1")
	if err != nil {
		t.Fatalf("CapturePane 不应报错: %v", err)
	}

	if out != ansiOutput {
		t.Errorf("期望完整保留 ANSI 转义符，实际输出: %q", out)
	}
	if !strings.Contains(out, "\x1b[32m") || !strings.Contains(out, "\x1b[33m") {
		t.Error("输出中缺失预期 ANSI 颜色控制字符")
	}
}

// 验证终端末尾 N 行截取辅助函数
func TestExtractTailLines(t *testing.T) {
	text := "line 1\nline 2\r\nline 3\nline 4\nline 5\n"

	tail3 := collector.ExtractTailLines(text, 3)
	if len(tail3) != 3 {
		t.Fatalf("期望提取 3 行，实际提取 %d 行", len(tail3))
	}
	if tail3[0] != "line 3" || tail3[1] != "line 4" || tail3[2] != "line 5" {
		t.Errorf("提取内容不符: %+v", tail3)
	}

	// 当总行数小于 N 时，返回全部有效行
	tail10 := collector.ExtractTailLines(text, 10)
	if len(tail10) != 5 {
		t.Fatalf("期望返回全部 5 行，实际返回 %d 行", len(tail10))
	}

	// 空字符串返回空切片
	emptyTail := collector.ExtractTailLines("", 5)
	if len(emptyTail) != 0 {
		t.Fatalf("空文本应返回空切片，实际返回 %d 行", len(emptyTail))
	}
}

// v1.REQ-002-AC-002：哈希计算与活跃状态更新跟踪
func TestPaneOutputTracker_HashAndActivityTracking(t *testing.T) {
	tracker := collector.NewPaneOutputTracker(50)
	t0 := time.Date(2026, 11, 11, 20, 0, 0, 0, time.UTC)
	t1 := t0.Add(5 * time.Second)

	// 第一次写入输出
	info1, changed1 := tracker.Update("%1", false, "hello world 1\n", t0)
	if !changed1 {
		t.Error("首次记录输出 changed 应为 true")
	}
	if info1.LastActiveAt != t0 {
		t.Errorf("首次活跃时间应为 t0，实际为 %v", info1.LastActiveAt)
	}
	if info1.LastHash == 0 {
		t.Error("首次写入应生成非零 64 位内容哈希")
	}

	// 内容未变，再次更新
	info2, changed2 := tracker.Update("%1", false, "hello world 1\n", t1)
	if changed2 {
		t.Error("内容无变动时 changed 应为 false")
	}
	if info2.LastActiveAt != t0 {
		t.Errorf("内容无变动时 LastActiveAt 应保持 t0，实际被更新为 %v", info2.LastActiveAt)
	}
	if info2.LastHash != info1.LastHash {
		t.Error("内容无变动时哈希应保持一致")
	}

	// 内容产生变动
	t2 := t1.Add(3 * time.Second)
	info3, changed3 := tracker.Update("%1", false, "hello world 2\n", t2)
	if !changed3 {
		t.Error("内容变动后 changed 应为 true")
	}
	if info3.LastActiveAt != t2 {
		t.Errorf("内容变动后 LastActiveAt 应刷新为 t2，实际为 %v", info3.LastActiveAt)
	}
	if info3.LastHash == info1.LastHash {
		t.Error("内容变动后哈希应改变")
	}
}

// v1.REQ-002-AC-003：当 pane 为 dead 状态时固化缓冲区并停止后续轮询
func TestPaneOutputTracker_DeadPaneFinalized(t *testing.T) {
	tracker := collector.NewPaneOutputTracker(50)
	t0 := time.Now()

	// 初始活跃
	tracker.Update("%2", false, "process running\n", t0)

	// 标记为已死
	infoDead, _ := tracker.Update("%2", true, "process running\n(pane dead)\n", t0.Add(time.Second))
	if !infoDead.IsFinalized {
		t.Error("dead pane 应标记为 IsFinalized = true")
	}

	// 再次更新已终结的 pane
	infoAfter, changed := tracker.Update("%2", true, "ignored new output", t0.Add(2*time.Second))
	if changed {
		t.Error("已终结的 pane 不应再触发 changed 变动")
	}
	if infoAfter.RawOutput != "process running\n(pane dead)\n" {
		t.Errorf("已终结的 pane 应保持终态内容，实际输出为: %s", infoAfter.RawOutput)
	}
}

// 验证 CapturePane 异常传递（如 server unavailable 与普通错误）
func TestCapturePane_Errors(t *testing.T) {
	t.Run("服务器不可用时返回 ErrTmuxUnavailable", func(t *testing.T) {
		mock := &MockTmuxExecutor{
			Outputs: map[string][]byte{
				"capture-pane": []byte("failed to connect to server"),
			},
			Errors: map[string]error{
				"capture-pane": errors.New("exit status 1"),
			},
		}

		c := collector.NewTmuxCollector(collector.Options{Executor: mock})
		_, err := c.CapturePane(context.Background(), "%1")
		if !errors.Is(err, collector.ErrTmuxUnavailable) {
			t.Errorf("期望返回 ErrTmuxUnavailable，实际返回: %v", err)
		}
	})

	t.Run("普通执行错误正确包装返回", func(t *testing.T) {
		mock := &MockTmuxExecutor{
			Outputs: map[string][]byte{
				"capture-pane": []byte("can't find pane: %99"),
			},
			Errors: map[string]error{
				"capture-pane": errors.New("exit status 1"),
			},
		}

		c := collector.NewTmuxCollector(collector.Options{Executor: mock})
		_, err := c.CapturePane(context.Background(), "%99")
		if err == nil || errors.Is(err, collector.ErrTmuxUnavailable) {
			t.Errorf("应返回常规执行错误，实际返回: %v", err)
		}
	})

	t.Run("NewPaneOutputTracker 缺省参数容错", func(t *testing.T) {
		tracker := collector.NewPaneOutputTracker(0)
		if tracker == nil {
			t.Fatal("NewPaneOutputTracker 不应返回 nil")
		}
	})
}

