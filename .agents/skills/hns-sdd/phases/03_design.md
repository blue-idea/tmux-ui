# STEP 3 · 技术方案（Design）

> **前置**：`docs/spec/versions/{{VERSION}}/requirements.md` 已定稿
> **输出文件**：
> - `docs/spec/versions/{{VERSION}}/design.md`（主文件）
> - `docs/spec/versions/{{VERSION}}/data.md`（数据库设计，按需）
> - `docs/spec/versions/{{VERSION}}/api.md`（接口设计，按需）
> **路径变量**：`{{VERSION}}` 取自 `docs/spec/README.md` 的 `current_version` 字段
> **跨 version 复用**：若设计沿用 vN 的 `data.md` / `api.md`，可显式引用 `docs/spec/versions/vN/data.md` 而不复制

---

## 目标

根据已定稿的需求文档和现有技术架构，完成技术方案设计，覆盖：项目结构、架构图、技术栈选型、安全性说明。

> **规则**：数据库细节写入 `data.md`，接口细节写入 `api.md`，`design.md` 中**只包含对它们的引用**，不重复编写。

---

## 执行步骤

1. 通读 `docs/spec/versions/{{VERSION}}/requirements.md`，识别核心数据实体和主要接口边界。
2. 使用 **Mermaid 图表**描述整体架构（必须包含）：
   - 系统架构图（组件关系）
   - 核心数据流图
3. 使用**表格**呈现技术选型决策，说明选型理由。
4. 数据库设计另存 `docs/spec/versions/{{VERSION}}/data.md`，接口设计另存 `docs/spec/versions/{{VERSION}}/api.md`。
5. 与用户确认后定稿，进入 STEP 4。

---

## design.md 必含章节

| 章节 | 内容说明 | 范围边界 |
|------|---------|---------|
| 项目结构 | 目录树（关键路径） | — |
| 架构概览 | Mermaid 系统架构图 | — |
| 技术选型 | 框架、ORM、状态管理等的选型表，含选型理由 | — |
| 测试工具选型 | 单元/集成/E2E 使用的测试框架（**仅框架名称**）| ⚠️ 只列框架，不写覆盖率目标和测试策略细节 |
| 安全性 | 认证方案、权限模型、密钥管理方式 | — |
| 数据库设计 | → 详见 `docs/spec/versions/{{VERSION}}/data.md`（如沿用 v1 则写 `docs/spec/versions/v1/data.md`）| — |
| 接口设计 | → 详见 `docs/spec/versions/{{VERSION}}/api.md` | — |

> **测试工具选型 vs 测试策略的边界**：
> - `design.md` 只写"使用什么框架"——例如"单元测试：Vitest；E2E：Playwright"。
> - 覆盖率目标、测试金字塔比例、风险等级决策、测试数据策略，全部留给 STEP 5 的 `test_strategy.md`。
> - 两份文档不得重复定义同一内容。如发现冲突，以 `test_strategy.md` 为准。

---

## 完成判定

- [ ] `docs/spec/versions/{{VERSION}}/design.md` 已创建，含 Mermaid 架构图
- [ ] `docs/spec/versions/{{VERSION}}/data.md` 已创建（如有数据库）
- [ ] `docs/spec/versions/{{VERSION}}/api.md` 已创建（如有 HTTP 接口）
- [ ] design.md 中的测试工具选型仅列框架名，未写覆盖率或策略细节
- [ ] 用户已确认定稿
