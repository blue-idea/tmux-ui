package history

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"tmux-ui/internal/config"
	"tmux-ui/internal/model"
)

var (
	// ErrSnapshotNotFound 快照不存在错误
	ErrSnapshotNotFound = errors.New("snapshot not found")
	// ErrPaneNotFound 目标 Pane 记录文件不存在错误
	ErrPaneNotFound = errors.New("pane history not found")
)

// SanitizePaneID 对 Pane ID 进行转义与路径安全净化，防止路径穿越攻击
func SanitizePaneID(paneID string) string {
	sanitized := strings.ReplaceAll(paneID, "%", "_")
	sanitized = strings.ReplaceAll(sanitized, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, "\\", "_")
	sanitized = strings.ReplaceAll(sanitized, "..", "__")
	sanitized = strings.ReplaceAll(sanitized, ".", "_")
	return sanitized
}

// JSONLStorage 基于结构化 JSON Lines 的本地追加存储引擎
type JSONLStorage struct {
	cfg       config.HistoryConfig
	locks     sync.Map // map[string]*sync.Mutex，单 Pane 独立锁
	globalMux sync.Mutex
}

// NewJSONLStorage 初始化 JSONL 本地持久化存储引擎
func NewJSONLStorage(cfg config.HistoryConfig) (*JSONLStorage, error) {
	if cfg.StorageDir == "" {
		cfg.StorageDir = "data/history"
	}
	if cfg.MaxSnapshotsPerPane <= 0 {
		cfg.MaxSnapshotsPerPane = 200
	}
	if cfg.MaxBytesPerPane <= 0 {
		cfg.MaxBytesPerPane = 10485760
	}

	if err := os.MkdirAll(cfg.StorageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create history storage directory: %w", err)
	}

	return &JSONLStorage{
		cfg: cfg,
	}, nil
}

// getPaneLock 获取指定 Pane 的互斥锁
func (s *JSONLStorage) getPaneLock(paneID string) *sync.Mutex {
	safeID := SanitizePaneID(paneID)
	val, _ := s.locks.LoadOrStore(safeID, &sync.Mutex{})
	return val.(*sync.Mutex)
}

// getFilePath 获取指定 Pane 的 .jsonl 文件路径
func (s *JSONLStorage) getFilePath(paneID string) string {
	safeID := SanitizePaneID(paneID)
	return filepath.Join(s.cfg.StorageDir, safeID+".jsonl")
}

// AppendSnapshot 将一条快照记录追加写入对应的 .jsonl 文件，并在超额时触发 FIFO 裁剪
func (s *JSONLStorage) AppendSnapshot(record *model.SnapshotRecord) error {
	if record == nil || record.PaneID == "" {
		return errors.New("invalid snapshot record: empty pane id")
	}

	if record.Timestamp <= 0 {
		record.Timestamp = time.Now().UnixMilli()
	}
	if record.SnapshotID == "" {
		record.SnapshotID = fmt.Sprintf("snap_%s_%d", SanitizePaneID(record.PaneID), record.Timestamp)
	}
	record.ByteLength = len(record.OutputBuffer)

	mu := s.getPaneLock(record.PaneID)
	mu.Lock()
	defer mu.Unlock()

	filePath := s.getFilePath(record.PaneID)

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot record: %w", err)
	}

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open history file: %w", err)
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		f.Close()
		return fmt.Errorf("failed to append snapshot to file: %w", err)
	}
	f.Close()

	// 检查是否超出最大记录数或大小并执行裁剪
	return s.pruneLocked(record.PaneID)
}

// GetSnapshot 查询特定 Pane 的单条历史快照详情（含完整 ANSI 终端内容）
func (s *JSONLStorage) GetSnapshot(paneID string, snapshotID string) (*model.SnapshotRecord, error) {
	mu := s.getPaneLock(paneID)
	mu.Lock()
	defer mu.Unlock()

	filePath := s.getFilePath(paneID)
	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrSnapshotNotFound
		}
		return nil, fmt.Errorf("failed to open history file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// 支持单行最大 4MB 缓冲区，防止大 ANSI 输出扫描溢出
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 4*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var record model.SnapshotRecord
		if err := json.Unmarshal(line, &record); err != nil {
			continue
		}
		if record.SnapshotID == snapshotID {
			return &record, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error while reading snapshots: %w", err)
	}

	return nil, ErrSnapshotNotFound
}

// ListSnapshots 查询指定 Pane 的快照轻量元数据列表，按时间正序排列
func (s *JSONLStorage) ListSnapshots(paneID string, since int64, limit int) ([]model.SnapshotMetadata, error) {
	mu := s.getPaneLock(paneID)
	mu.Lock()
	defer mu.Unlock()

	if limit <= 0 {
		limit = 100
	}

	filePath := s.getFilePath(paneID)
	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.SnapshotMetadata{}, nil
		}
		return nil, fmt.Errorf("failed to open history file: %w", err)
	}
	defer f.Close()

	var result []model.SnapshotMetadata
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 4*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var meta model.SnapshotMetadata
		if err := json.Unmarshal(line, &meta); err != nil {
			continue
		}

		if meta.Timestamp >= since {
			result = append(result, meta)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error while listing snapshots: %w", err)
	}

	// 如果超出 limit，截取最新的 limit 条
	if len(result) > limit {
		result = result[len(result)-limit:]
	}

	return result, nil
}

// pruneLocked 在持有 Pane 锁的前提下执行 FIFO 滚动淘汰
func (s *JSONLStorage) pruneLocked(paneID string) error {
	filePath := s.getFilePath(paneID)

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	var lines [][]byte
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 4*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) > 0 {
			// 必须深拷贝 bytes 避免 scanner 缓冲区重用污染
			cp := make([]byte, len(line))
			copy(cp, line)
			lines = append(lines, cp)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// 判定是否需要裁剪（行数超限或字节大小超限）
	needsPruning := len(lines) > s.cfg.MaxSnapshotsPerPane || fileInfo.Size() > s.cfg.MaxBytesPerPane
	if !needsPruning {
		return nil
	}

	// 按 FIFO 截断，仅保留最新的 MaxSnapshotsPerPane 条
	keepCount := s.cfg.MaxSnapshotsPerPane
	if len(lines) > keepCount {
		lines = lines[len(lines)-keepCount:]
	}

	// 原子替换文件：先写入临时文件，再重命名覆盖
	tmpFile := filePath + ".tmp"
	tf, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create temp file for pruning: %w", err)
	}

	for _, line := range lines {
		if _, err := tf.Write(append(line, '\n')); err != nil {
			tf.Close()
			os.Remove(tmpFile)
			return fmt.Errorf("failed to write to temp prune file: %w", err)
		}
	}
	tf.Close()

	if err := os.Rename(tmpFile, filePath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to atomic rename prune file: %w", err)
	}

	return nil
}
