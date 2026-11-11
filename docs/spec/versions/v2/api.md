# 接口设计与实时通信协议（API & WebSocket Specification）

> 文件路径：`docs/spec/versions/v2/api.md`
> 适用版本：v2 · 生成日期：2026-11-12
> 引用自：`docs/spec/versions/v2/design.md`
> 继承自：`docs/spec/versions/v1-finished/api.md`（完整保留 v1 原有端点并扩展 v2 历史回放能力）

---

## 1. 通用约定

### 1.1 基础路径与鉴权
- **Base URL**：`http://127.0.0.1:8080/api/v1`
- **WebSocket URL**：`ws://127.0.0.1:8080/ws`
- **鉴权方式**：
  - HTTP Header：`Authorization: Bearer <AUTH_TOKEN>`
  - WebSocket 握手：URL Query 参数 `ws://127.0.0.1:8080/ws?token=<AUTH_TOKEN>`
  - 若服务端未配置 `--token`，则自动放行所有请求。

### 1.2 统一响应信封
```json
// 成功
{
  "code": 0,
  "message": "success",
  "data": { ... }
}

// 错误
{
  "error": {
    "code": "SNAPSHOT_NOT_FOUND",
    "message": "Snapshot 'snap__0_12345' does not exist for pane '%0'.",
    "details": {
      "pane_id": "%0",
      "snapshot_id": "snap__0_12345"
    }
  }
}
```

---

## 2. v2 新增 HTTP REST 接口

### 2.1 获取 Pane 历史快照元数据列表
- **路径**：`GET /api/v1/panes/{id}/snapshots`
- **描述**：拉取指定 Pane 的历史快照时间轴元数据列表（轻量级，不包含 ANSI 输出大体积文本），用于前端时间轴滑块渲染刻度与事件节点。
- **URL 参数**：
  - `{id}`：Pane ID，支持 `%0`、URL 编码 `%250` 或裸数字 `0`。
- **Query 参数**：
  - `since`（可选，int64）：起始毫秒时间戳（仅返回晚于该时刻的快照）。
  - `limit`（可选，int，默认 100，最大 200）：最大返回快照数。
- **成功响应（HTTP 200）**：
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "snapshot_id": "snap__0_1731380400000_1a2b",
      "pane_id": "%0",
      "timestamp": 1731380400000,
      "event_type": "state_transition",
      "old_state": "idle",
      "new_state": "working",
      "summary": "State changed from idle to working",
      "byte_length": 8192
    },
    {
      "snapshot_id": "snap__0_1731380460000_3c4d",
      "pane_id": "%0",
      "timestamp": 1731380460000,
      "event_type": "state_transition",
      "old_state": "working",
      "new_state": "waiting_input",
      "summary": "State changed from working to waiting_input",
      "byte_length": 14250
    }
  ]
}
```

---

### 2.2 获取特定历史快照详情与 ANSI 内容
- **路径**：`GET /api/v1/panes/{id}/snapshots/{snapshot_id}`
- **描述**：根据快照 ID 获取当时保存的完整 ANSI 终端内容，用于在终端查看器中回溯渲染。
- **URL 参数**：
  - `{id}`：Pane ID（同上宽容路由规则）。
  - `{snapshot_id}`：快照唯一标识符。
- **成功响应（HTTP 200）**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "snapshot_id": "snap__0_1731380460000_3c4d",
    "pane_id": "%0",
    "session_name": "main",
    "window_id": "@0",
    "timestamp": 1731380460000,
    "event_type": "state_transition",
    "old_state": "working",
    "new_state": "waiting_input",
    "summary": "State changed from working to waiting_input",
    "byte_length": 14250,
    "output_buffer": "\u001b[32m[Claude Code]\u001b[0m Running tests...\n\u001b[33mDo you want to proceed? (y/n)\u001b[0m",
    "tail_lines": [
      "Running tests...",
      "Do you want to proceed? (y/n)"
    ]
  }
}
```
- **错误响应（HTTP 404）**：
```json
{
  "error": {
    "code": "SNAPSHOT_NOT_FOUND",
    "message": "Snapshot 'snap__0_99999' not found."
  }
}
```

---

## 3. WebSocket 协议增强

### 3.1 增量事件广播：`snapshot_created`
当 History Engine 记录下新的快照（无论是状态跃迁还是 Checkpoint）时，向当前订阅该 Pane 的 WebSocket 客户端广播快照元数据，驱动前端时间轴滑块动态扩展刻度：

```json
{
  "event": "snapshot_created",
  "timestamp": 1731380460000,
  "data": {
    "snapshot_id": "snap__0_1731380460000_3c4d",
    "pane_id": "%0",
    "timestamp": 1731380460000,
    "event_type": "state_transition",
    "old_state": "working",
    "new_state": "waiting_input",
    "summary": "State changed from working to waiting_input",
    "byte_length": 14250
  }
}
```
