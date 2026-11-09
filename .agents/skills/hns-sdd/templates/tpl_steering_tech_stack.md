# 锁定技术栈（Tech Stack）

> 文件路径：`docs/spec/steering/tech-stack.md`
> 创建步骤：STEP 1（首次立规矩时生成）
> ⚠️ 本文件**不随 version 变更**。技术栈变更属于重大决策，应升级 version 并记录原因

---

## 主语言

- 语言：{{LANGUAGE}}
- 版本：{{LANGUAGE_VERSION}}

## 运行时 / 框架

- 框架：{{FRAMEWORK}}
- 运行时：{{RUNTIME_VERSION}}

## 包管理

- 包管理器：**pnpm**（统一，不混用 npm / yarn）
- 锁文件：`pnpm-lock.yaml` 必须提交到版本控制
- 禁止使用任何 beta / rc / alpha 版本

## TypeScript 规范（如适用）

- 启用严格模式（`strict: true`）
- 禁止使用 `any` 和 `unknown` 类型
- 禁止使用 `as` 断言
- 禁止编写带 `as any` 或 lint 抑制注释的不安全代码
- 任务完成后自动执行 `pnpm exec tsc --noEmit` 检查

## 数据库

- 数据库：{{DATABASE}}（版本：{{DATABASE_VERSION}}）
- ORM / 查询工具：{{ORM}}
- 迁移工具：{{MIGRATION_TOOL}}

## 测试框架

| 测试类型 | 框架 | 备注 |
|---------|------|------|
| 单元 / 组件 | Vitest + Testing Library | — |
| API | `curl` / Supertest | 直接执行，保留真实输出 |
| Web E2E | Playwright MCP | 需截图留证 |
| 移动端 E2E | Maestro CLI | 需截图留证 |
| 性能 | k6 / Locust | 按需 |
| 安全 | ZAP / 手动 | OWASP Top 10 |

## CI / CD

- 平台：GitHub Actions
- 阶段：lint → type-check → unit → integration → E2E → build
- 强制检查：{{ADDITIONAL_CI_RULES}}