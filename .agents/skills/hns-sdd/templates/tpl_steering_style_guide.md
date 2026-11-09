# 命名 / 错误格式 / 日志（Style Guide）

> 文件路径：`docs/spec/steering/style-guide.md`
> 创建步骤：STEP 1（首次立规矩时生成）
> ⚠️ 本文件**不随 version 变更**

---

## 命名规范

### 文件与目录

| 类型 | 规范 | 示例 |
|------|------|------|
| 组件文件 | PascalCase | `UserProfile.tsx` |
| 工具函数文件 | camelCase | `formatDate.ts` |
| 常量文件 | camelCase | `constants.ts` |
| 测试文件 | 与被测文件同名 + `.test` / `.spec` | `calculatePrice.test.ts` |
| 目录 | kebab-case | `user-profile/` |

### 代码标识符

| 类型 | 规范 | 示例 |
|------|------|------|
| 变量 | camelCase | `userName` |
| 常量 | UPPER_SNAKE_CASE | `MAX_RETRY_COUNT` |
| 函数 | camelCase，动词开头 | `calculateTotal()` |
| 类 | PascalCase | `OrderProcessor` |
| 接口 / 类型 | PascalCase | `UserProfile` |
| 枚举 | PascalCase，值 UPPER_SNAKE_CASE | `OrderStatus.PENDING` |
| 数据库表 | snake_case 复数 | `user_profiles` |
| 数据库字段 | snake_case | `created_at` |

---

## 错误格式

所有用户可见的错误必须满足：

1. **可读**：面向最终用户，不暴露堆栈或 SQL
2. **可定位**：包含 `code`（机器可识别）和 `message`（人可读）
3. **可行动**：告知用户下一步该做什么（如"请检查邮箱后重试"）

### 错误响应 Schema

```json
{
  "error": {
    "code": "EMAIL_ALREADY_EXISTS",
    "message": "该邮箱已被注册",
    "details": { "field": "email" },
    "trace_id": "abc-123-def"
  }
}
```

### HTTP 状态码使用

| 场景 | 状态码 |
|------|-------|
| 创建成功 | 201 |
| 查询/更新成功 | 200 |
| 请求格式错误 | 400 |
| 认证失败 | 401 |
| 权限不足 | 403 |
| 资源不存在 | 404 |
| 业务规则违反 | 422 |
| 速率限制 | 429 |
| 依赖宕机 | 503 |

---

## 日志格式

### 结构化字段

所有日志输出 JSON，包含：

| 字段 | 说明 |
|------|------|
| `timestamp` | ISO 8601（含时区） |
| `level` | `debug` / `info` / `warn` / `error` |
| `service` | 服务名 |
| `trace_id` | 链路追踪 ID |
| `message` | 一句话描述 |
| `context` | 业务上下文（user_id、order_id 等） |

### 等级使用

| 等级 | 何时使用 |
|------|---------|
| `debug` | 详细调试信息，仅本地/开发 |
| `info` | 关键业务事件（用户登录、订单创建） |
| `warn` | 异常但可恢复（重试、降级） |
| `error` | 失败需人工介入 |

### 禁止事项

- ⛔ 禁止在生产日志中输出密码、token、PII
- ⛔ 禁止使用 `console.log` 替代正式 logger
- ⛔ 禁止吞掉异常（`catch {}` 空块）