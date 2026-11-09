# 项目协作与测试门槛（Project Rules）

> 文件路径：`docs/spec/steering/project-rules.md`
> 创建步骤：STEP 1（首次立规矩时生成）
> ⚠️ 本文件**不随 version 变更**，只在项目级做出调整时更新

---

## 沟通规范

1. 所有内容使用**中文**交流，包括代码注释和测试描述。
2. 仅以下内容使用英文：用户界面文案、错误提示消息、日志消息、第三方 API 字段名。
3. 遇到不清楚、歧义、缺失内容时，立即向用户提问，不得自行假设。
4. 表达清晰、简洁、技术准确。

### 代码语言规范速查

| 内容 | 语言 | 示例 |
|------|------|------|
| 代码注释 | 中文 | `// 计算折扣后的最终价格` |
| 测试描述（`test()`/`it()`/`describe()`） | 中文 | `test('提交空邮箱时返回错误', ...)` |
| 函数名 / 变量名 | 英文（代码惯例） | `calculateDiscount()` |
| 错误消息字符串 | 英文 | `'Email is required'` |
| UI 文案字符串 | 英文 | `'Submit'` |
| Git commit message | 中文或英文均可，项目内保持一致 | — |

---

## Git 协作规范

### 主分支

- 主分支：`main`（不存在则 `master`）
- 默认不强制 `push` 远程，除非项目规范明确要求

### 任务分支

- 每个 TASK 独立一条分支：`feat/TASK-XXX` 或 `feat/TASK-XXX-short-slug`
- 验收通过后再合并到主分支；冲突时先报告用户，禁止 `--force`

### Commit

- 单次 commit 对应一个原子改动（一个 TASK 的一次提交）
- message 格式：`TASK-XXX: <简要说明>`

---

## 测试门槛

- 每个 TASK 必须在 PR 前通过：
  - `pnpm exec tsc --noEmit` 零 error
  - `pnpm exec eslint src --max-errors=0` 零 error
  - `pnpm test --run` 100% 通过
  - `pnpm build` 零 error
- 无法执行的用例标记 `BLOCKED` 并说明原因；**禁止伪造 PASS**

### TDD 铁律

```
⛔ 没有失败的测试，禁止写任何生产代码
```

完整协议见 `hns-sdd/phases/tdd.md`（Agent 在 STEP 6 执行时加载）。

---

## 包管理器

项目统一使用 **pnpm**。所有命令示例均使用 pnpm，不得混用 npm 或 yarn。

---

## 环境变量

- 所有密钥与敏感信息统一放在 `.env` / `.env.local`
- 代码中通过 `process.env.XXX` 引用，禁止硬编码
- `.env` 必须加入 `.gitignore`
- 测试账号统一从 `docs/spec/info.md` 获取

---

## 工程化要求

1. **严禁重复**：任何重复逻辑必须抽象为可复用单元
2. **检查现有代码**：先分析项目中是否已有类似实现
3. **模块化设计**：将功能拆分为独立、可复用的模块
4. **统一接口**：使用一致的参数格式和返回值结构
5. **配置集中管理**：所有常量和配置项统一放在 `config/` 目录
6. **优先封装**：将相关功能和数据封装在职责明确的模块中