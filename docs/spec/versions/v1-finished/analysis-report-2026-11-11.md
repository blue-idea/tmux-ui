# 规格审查报告（Analysis Report）

> 文件路径：`docs/spec/versions/v1/analysis-report-2026-11-11.md`
> 适用版本：v1
> 创建步骤：STEP 4（规格审查与分析）
> 审查日期：2026-11-11
> 审查人：Antigravity Agent
> 状态：审查完成（Archived）

---

## 1. 审查范围

| 文件 | 版本 / 最后修改日期 |
|------|------------------|
| `docs/spec/versions/v1/requirements.md` | v1 · 2026-11-11（已更新） |
| `docs/spec/versions/v1/design.md` | v1 · 2026-11-11 |
| `docs/spec/versions/v1/data.md` | v1 · 2026-11-11（已更新） |
| `docs/spec/versions/v1/api.md` | v1 · 2026-11-11（已更新） |

---

## 2. 审查结论摘要

- **发现问题数**：P0 0 条 · P1 2 条 · P2 1 条
- **已更新文件**：`requirements.md`、`data.md`、`api.md`
- **需用户决策**：2 条（已全部确认闭环）
- **结论**：✅ 审查通过，可正式进入 STEP 5（拆分任务）

---

## 3. 问题列表

| # | 优先级 | 关联需求 | 维度 | 问题描述 | 处理方式 | 状态 |
|---|:------:|---------|------|---------|---------|:----:|
| 1 | P1 | v1.REQ-001 | 前置条件与约束 | 需明确多 tmux server 或自定义 socket（`-L` / `-S`）的配置与传参方式 | 经用户确认，采取 CLI 参数 + YAML 双重支持（CLI 优先） | 已解决 |
| 2 | P1 | v1.REQ-005 | 跨文档冲突 / 易用性 | `GET /api/v1/panes/{id}` 中，tmux pane 原生 ID 以 `%` 开头，直接传 `%` 会引起 URL 解析问题 | 经用户确认，采取宽容路由解析：支持裸数字 `1` 自动补 `%1` 与 `%251` 解码 | 已解决 |
| 3 | P2 | v1.REQ-002 | 非功能性需求 | 终端输出缓冲区达到上限（512KB）时的截断策略与内存清理周期 | 已在 data.md 与 design.md 明确采用头部滑动丢弃机制及 10 分钟 TTL 僵尸清理 | 已解决 |

---

## 4. 澄清问题记录

### 澄清 1 · 自定义 Tmux Socket 传参方式【v1.REQ-001】

**问题**：当宿主机存在多个 tmux server 或使用自定义 socket 时，后端程序应如何获取并指定 `-L socket名` 或 `-S 路径`？

**选项**：
- A. 仅支持命令行启动参数（`--socket-name` / `--socket-path`）
- B. 仅支持在 `config/agents.yaml` 配置文件中配置
- C. 双重支持：支持 CLI 命令行启动参数与配置文件配置，且命令行参数优先级高于配置文件（推荐）
- D. 其他（请说明）

**用户回复**：C（双重支持：支持 CLI 命令行启动参数与配置文件配置，且命令行参数优先级高于配置文件）

**处理动作**：
1. 在 `v1.REQ-001` 中新增验收标准 `v1.REQ-001-AC-005`，明确 CLI 参数与配置文件联合支持机制。
2. 在 `data.md` 的 `collector:` 配置结构中追加 `socket_name` 与 `socket_path` 字段。
- **更新文件**：`requirements.md`, `data.md`

---

### 澄清 2 · Pane ID URL 路由解析兼容性【v1.REQ-005】

**问题**：在 HTTP 接口 `GET /api/v1/panes/{id}` 中，针对带 `%` 的 tmux pane ID（例如 `%1`），系统应如何处理路由参数？

**选项**：
- A. 宽容兼容：同时支持不带 `%` 的裸 ID（如 `1`，服务端自动补齐为 `%1`）和已 URL 转义的 `%251`（推荐）
- B. 严格转义：强制要求前端将 `%` 编码为 `%251`，否则报 400 Bad Request
- C. 改用 Query 参数：调整接口为 `GET /api/v1/pane?id=%1` 避免路径编码陷阱
- D. 其他（请说明）

**用户回复**：A（宽容兼容：同时支持不带 `%` 的裸 ID 和已 URL 转义的 `%251`）

**处理动作**：
1. 更新 `v1.REQ-005-AC-002`，规范化参数解析描述。
2. 更新 `api.md` 接口契约，注明路由参数归一化算法（裸数字补齐 `%`，转义字符先 URL 解码）。
- **更新文件**：`requirements.md`, `api.md`

---

## 5. 文档变更清单

| 文件 | 变更内容摘要 | 变更类型 |
|------|------------|---------|
| `docs/spec/versions/v1/requirements.md` | 新增 `v1.REQ-001-AC-005`（Socket 多源传参）；更新 `v1.REQ-005-AC-002`（宽容路由） | 修改 |
| `docs/spec/versions/v1/data.md` | `collector` 配置模型新增 `socket_name` 与 `socket_path` | 修改 |
| `docs/spec/versions/v1/api.md` | `GET /api/v1/panes/{id}` 接口参数说明补充宽容路由兼容定义 | 修改 |
| `docs/spec/versions/v1/analysis-report-2026-11-11.md` | 完成审查发现问题记录、用户决策归档并标记完成 | 新增/归档 |

---

## 6. 结论

所有 7 个维度的审查已全部完成，澄清问题均已得到用户决策闭环，相关规格文档已同步回写完毕。
无遗留 P0/P1 问题，规格体系完整、一致、可追溯，**正式通过审查，准予进入 STEP 5 · 拆分任务（Tasks）**。
