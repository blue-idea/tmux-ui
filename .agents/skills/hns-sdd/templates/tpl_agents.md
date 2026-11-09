# AGENTS.md — 项目级 Agent 入口

> 文件路径：`docs/AGENTS.md`
> 创建步骤：STEP 1（首次立规矩时生成）
> 本文件告诉 Agent 项目当前活跃的 SDD version 与恒定规则位置

---

## 入口说明

本项目遵循 `hns-sdd` 技能流程。Agent 进入项目后，**先读取本文件**，再读取 `docs/spec/README.md` 取得 `current_version`，随后按 SDD 阶段执行。

---

## 必读文件（按顺序）

1. `docs/spec/README.md` — 当前活跃 version（`current_version`）
2. `docs/spec/steering/project-rules.md` — 协作 / Git / 测试门槛
3. `docs/spec/steering/tech-stack.md` — 锁定技术栈
4. `docs/spec/steering/style-guide.md` — 命名 / 错误格式 / 日志
5. `docs/spec/versions/{{VERSION}}/constitution.md` — 项目宪法（vN+1 沿用 v1）
6. `docs/spec/versions/{{VERSION}}/requirements.md` — 当前 version 需求（vN 阶段 STEP 2+ 必读）

> `{{VERSION}}` 替换为 `docs/spec/README.md` 中 `current_version` 的值。

---

## 触发词

在本项目目录中，可使用以下触发词唤起 SDD 流程：

- `按 SDD 流程进入 STEP N`
- `按 SDD 流程为当前 version 编写 requirements.md`
- `执行 TASK-XXX`
- `按规格回退协议处理`

---

## 跨 version 引用规范

- REQ/AC 编号建议带 version 前缀（如 `v1.REQ-001`、`v2.REQ-001`）
- constitution 引用 `docs/spec/versions/v1/constitution.md`（v1 是唯一权威源）
- data.md / api.md 跨 version 引用时使用 `docs/spec/versions/v1/data.md` 显式路径

---

## 与 hns-sdd 技能的关系

本文件不是 skill 包的一部分，是项目级入口。Skill 包位于 `hns-sdd/`（或 Codex/Claude skills 目录），由 Agent 按需加载 `SKILL.md` 与 `phases/*`。