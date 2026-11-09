# STEP 1 · 立规矩（Constitution）

> **前置**：已完成 STEP 0（通用规范已加载）
> **输出文件**：
> - `docs/AGENTS.md`（项目级 Agent 入口，仅首次创建）
> - `docs/spec/README.md`（版本索引，含 `current_version`，仅首次创建）
> - `docs/spec/steering/project-rules.md` / `tech-stack.md` / `style-guide.md`（仅首次创建）
> - `docs/spec/versions/{{VERSION}}/constitution.md`（v1 首次生成；vN+1 引用 v1）
> **模板参考**：`templates/tpl_constitution.md`
> **路径变量**：`{{VERSION}}` 取自 `docs/spec/README.md` 的 `current_version` 字段；首次进入 STEP 1 取 `v1`
>
> > **v2+ 注意事项**：`constitution.md`、`steering/`、`test_strategy.md`、`info.md` 为项目级资产，v2+ 不重复生成。v2+ 的 per-version 核心文件为 `requirements.md`、`tasks.md`、`traceability.md`。

---

## 目标

为项目建立**技术宪法**——一份覆盖技术栈约束、工程化要求、沟通规范、安全与隐私、一致性要求的基础文件。

> constitution **不涉及具体功能逻辑**，只约束"如何做事"。

---

## 执行步骤

1. 加载 `templates/tpl_constitution.md`，以其为骨架开始填写。
2. 与用户确认以下核心维度（逐条确认，不得跳过）：

| 维度 | 需确认的内容 |
|------|------------|
| 技术栈 | 主语言、框架、运行时版本 |
| 包管理 | 使用 npm / pnpm / yarn，版本锁定策略 |
| TypeScript | 严格模式开关，`any` / `as` 使用约束 |
| 编码安全 | RLS、Rate Limiting、API 密钥管理方式 |
| 质量基线 | Lint 规则、Git Hooks、CI/CD 配置 |
| 测试策略 | 单元测试框架、E2E 框架 |
| 文档结构 | `docs/spec/versions/{{VERSION}}/` 目录下各文件职责 |

3. 所有维度确认后：

   - **首次进入 STEP 1**（无既有目录）：
     1. 创建 `docs/spec/versions/v1/` 目录。
     2. 生成 `docs/spec/README.md`，写入：
        ```yaml
        current_version: v1
        versions:
          - v1
        ```
     3. 生成 `docs/spec/steering/project-rules.md`、`tech-stack.md`、`style-guide.md`（从 constitution 拆出"如何做事"部分，避免重复）。
     4. 生成 `docs/AGENTS.md`，指向 `docs/spec/README.md` 与 `docs/spec/steering/`。
     5. 生成 `docs/spec/versions/v1/constitution.md`。
   - **vN+1**（`constitution.md` 已存在 v1）：沿用 v1，**不重新生成**。如需微调，沿用项目"宪法修订"流程（重大变更 → 升级 version；微调 → 直接修改 v1 并在 `docs/spec/versions/v1/constitution.md` 顶部记录变更日志）。

4. 与用户二次确认文件内容，定稿后**禁止再修改**（如需变更须显式通知并记录原因）。

---

## 完成判定

- [ ] `docs/spec/versions/{{VERSION}}/constitution.md` 文件已创建（或确认沿用 v1）
- [ ] `docs/spec/steering/*.md` 已首次创建
- [ ] `docs/spec/README.md` 已首次创建
- [ ] `docs/AGENTS.md` 已首次创建
- [ ] 用户已明确确认内容定稿
- [ ] 文件中无占位符（`[TODO]` / `...` 等）残留
