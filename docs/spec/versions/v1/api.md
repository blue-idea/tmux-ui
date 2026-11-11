# 接口设计与实时通信协议（API & WebSocket Specification）

> 文件路径：`docs/spec/versions/v1/api.md`
> 适用版本：v1 · 生成日期：2026-11-11
> 引用自：`docs/spec/versions/v1/design.md`

---

## 1. 通用约定

### 1.1 基础路径与鉴权
- **Base URL**：`http://127.0.0.1:8080/api/v1`
- **WebSocket URL**：`ws://127.0.0.1:8080/ws`
- **鉴权方式**：
  - HTTP 请求头：`Authorization: Bearer <AUTH_TOKEN>`
  - WebSocket 握手：URL 查询参数 `ws://127.0.0.1:8080/ws?token=<AUTH_TOKEN>` 或 `Sec-WebSocket-Protocol`
  - 若服务启动时未开启 Token 保护（默认模式），则无需提供 Token。

### 1.2 统一响应结构

#### 成功响应（HTTP 200/201）
```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

#### 错误响应（HTTP 4xx/5xx）
```json
{
  "error": {
    "code": "PANE_NOT_FOUND",
    "message": "The requested pane '%99' does not exist.",
    "details": {
      "pane_id": "%99"
    }
  }
}
```

---

## 2. HTTP REST 接口列表

### 2.1 健康检查与系统信息
- **请求**：`GET /api/v1/health`
- **权限**：无需认证
- **响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "healthy",
    "tmux_version": "tmux 3.4",
    "active_panes": 4,
    "uptime_seconds": 3600
  }
}
```

### 2.2 获取所有 Pane 摘要列表（总览卡片墙）
- **请求**：`GET /api/v1/panes`
- **查询参数**：
  - `status` (可选)：按状态过滤，如 `status=waiting_input`
  - `session` (可选)：按 tmux session 过滤
- **响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "pane_id": "%1",
      "session_name": "agent-dev",
      "window_id": "@0",
      "window_name": "claude-runner",
      "pane_pid": 12345,
      "current_command": "node",
      "agent_type": "claude-code",
      "state": "waiting_input",
      "last_active_at": "2026-11-11T19:20:00Z",
      "idle_seconds": 12,
      "is_dead": false,
      "preview_lines": [
        "Thinking about refactoring...",
        "Do you want to proceed? (y/n)"
      ]
    }
  ]
}
```

### 2.3 获取单个 Pane 完整详情（终端交互页）
- **请求**：`GET /api/v1/panes/{id}`（例如 `GET /api/v1/panes/1` 或 `GET /api/v1/panes/%251`）
- **路径参数**：`id` 为 tmux pane 标识符。服务端支持宽容解析：若传入裸数字 `1` 自动补全为 `%1`；若传入 URL 编码的 `%251` 自动解码为 `%1`。
- **响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "pane_id": "%1",
    "session_name": "agent-dev",
    "window_name": "claude-runner",
    "agent_type": "claude-code",
    "state": "waiting_input",
    "raw_output_ansi": "\u001b[32m[Agent]\u001b[0m Refactoring done.\nDo you want to proceed? (y/n)",
    "timeline": [
      {
        "id": "evt-01",
        "pane_id": "%1",
        "timestamp": "2026-11-11T19:18:00Z",
        "from_state": "idle",
        "to_state": "working",
        "trigger_rule": "matched process claude"
      },
      {
        "id": "evt-02",
        "pane_id": "%1",
        "timestamp": "2026-11-11T19:19:45Z",
        "from_state": "working",
        "to_state": "waiting_input",
        "trigger_rule": "regex matched 'Do you want to proceed?'"
      }
    ],
    "processes": [
      {
        "pid": 12345,
        "ppid": 1,
        "command": "/bin/zsh",
        "children": [
          {
            "pid": 12380,
            "ppid": 12345,
            "command": "node /usr/local/bin/claude"
          }
        ]
      }
    ]
  }
}
```

### 2.4 获取当前加载的 Agent 识别规则
- **请求**：`GET /api/v1/config/agents`
- **响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "agents": [
      {
        "name": "claude-code",
        "match_process": ["claude", "node"],
        "states": {
          "working": { "tail_regex": "esc to interrupt" },
          "waiting_input": { "tail_regex": "Do you want to proceed\\?" }
        }
      }
    ]
  }
}
```

---

## 3. WebSocket 实时推流协议

客户端通过 `ws://127.0.0.1:8080/ws` 连接到服务端，建立全双工实时事件通道。

### 3.1 握手与初始化快照
连接建立成功后，服务端立即向客户端推送 `snapshot` 全量事件：

```json
{
  "type": "snapshot",
  "timestamp": "2026-11-11T19:25:00Z",
  "payload": {
    "panes": [ ...PaneSummary... ]
  }
}
```

### 3.2 服务端推送事件类型（Server -> Client）

| 事件类型 (`type`) | 触发时机 | Payload 说明 |
|------------------|---------|-------------|
| `snapshot` | 客户端首次连入 | 所有 Pane 的当前状态数组 |
| `pane_added` | tmux 检测到新 Pane 创建 | 新增的 `PaneSummary` |
| `pane_removed` | tmux 检测到 Pane 销毁 | `{ "pane_id": "%1" }` |
| `state_changed` | 智能体状态产生变迁 | `{ "pane_id": "%1", "from": "working", "to": "waiting_input", "reason": "...", "timestamp": "..." }` |
| `output_diff` | Pane 终端产生新输出 | `{ "pane_id": "%1", "delta_ansi": "...", "last_active_at": "..." }` |

#### `state_changed` 示例
```json
{
  "type": "state_changed",
  "timestamp": "2026-11-11T19:25:10Z",
  "payload": {
    "pane_id": "%1",
    "from_state": "working",
    "to_state": "waiting_input",
    "trigger_rule": "regex matched 'Do you want to proceed?'",
    "preview_snippet": "Do you want to proceed? (y/n)"
  }
}
```

### 3.3 客户端上行控制消息（Client -> Server）

```json
// 心跳保活
{ "action": "ping" }

// 响应服务端保活
{ "type": "pong" }

// 详情页专注订阅指定 Pane 输出流（减少非活动 pane 推流带宽）
{
  "action": "subscribe_pane",
  "pane_id": "%1"
}
```

---

## 4. 错误码定义字典

| 错误码 (`code`) | HTTP 状态码 | 英文描述 | 解决方案 |
|----------------|------------|---------|---------|
| `UNAUTHORIZED` | 401 | Authentication token is invalid or missing | 检查请求头中的 Bearer Token |
| `TMUX_UNAVAILABLE` | 503 | Tmux server is not running or socket is inaccessible | 检查宿主机 tmux 服务是否已启动 |
| `PANE_NOT_FOUND` | 404 | Target tmux pane does not exist | 确认目标 pane 是否已被关闭 |
| `INVALID_PARAMETER` | 400 | Query or path parameters are invalid | 核对接口参数格式 |
| `INTERNAL_ERROR` | 500 | Unexpected internal server error | 查看服务器结构化日志 |

