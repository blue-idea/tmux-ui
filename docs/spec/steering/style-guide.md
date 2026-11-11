# 命名 / 错误格式 / 日志（Style Guide）

> 文件路径：`docs/spec/steering/style-guide.md`
> 创建步骤：STEP 1（首次立规矩时生成）
> ⚠️ 本文件**不随 version 变更**

---

## 命名规范

### 1. Go 后端

| 类型 | 规范 | 示例 |
|------|------|------|
| 包名 | 纯小写单个单词或简写，无下划线 | `collector`, `detector`, `store` |
| 导出函数 / 结构体 / 接口 | PascalCase | `CollectPanes()`, `AgentState`, `PaneDetector` |
| 私有函数 / 变量 | camelCase | `parsePaneFormat()`, `lastOutputHash` |
| 常量 | PascalCase 或 全大写 | `DefaultPollInterval`, `StateWorking` |
| 测试文件 | 源码文件同名 + `_test.go` | `collector_test.go` |

### 2. React / TypeScript 前端

| 类型 | 规范 | 示例 |
|------|------|------|
| 组件文件 | PascalCase | `PaneCard.tsx`, `TerminalViewer.tsx` |
| Hook 文件 | camelCase 以 `use` 开头 | `useTmuxState.ts` |
| 工具函数文件 | camelCase | `ansiParser.ts` |
| 测试文件 | 与被测文件同名 + `.test.tsx` / `.spec.ts` | `PaneCard.test.tsx` |
| 样式 / 目录 | kebab-case | `components/pane-card/` |

---

## 接口与错误响应格式

所有 HTTP API 响应统一遵循标准包装：

### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

### 错误响应

```json
{
  "error": {
    "code": "TMUX_SOCKET_NOT_FOUND",
    "message": "Tmux socket not found or tmux server is not running.",
    "details": {
      "socket_path": "/tmp/tmux-501/default"
    }
  }
}
```

### 常用 HTTP 状态码

| 场景 | 状态码 |
|------|-------|
| 成功获取 | 200 |
| 请求参数错误 | 400 |
| 未授权 / Token 缺失 | 401 |
| 权限拒绝 | 403 |
| 目标 Pane 不存在 | 404 |
| 内部执行错误（如 tmux 命令报错） | 500 |

---

## 日志规范

Go 后端使用结构化日志（如 `log/slog`），输出 JSON 格式：

| 字段 | 类型 | 说明 |
|------|------|------|
| `time` | string | ISO 8601 时间戳 |
| `level` | string | `DEBUG`, `INFO`, `WARN`, `ERROR` |
| `msg` | string | 简明英文信息 |
| `pane_id` | string | 发生事件的 tmux pane 标识（如 `%1`，可选） |
| `error` | string | 错误详细信息（可选） |

### 禁止事项

- ⛔ 严禁在日志中泄露用户敏感环境变量或用户认证 Token
- ⛔ 严禁使用未格式化的全局 `fmt.Println` 充当生产日志

