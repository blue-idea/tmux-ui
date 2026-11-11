# docs/spec/README.md — 版本索引

> 文件路径：`docs/spec/README.md`
> 创建步骤：STEP 1（首次立规矩时生成）
> 本文件记录当前活跃 version，由 Agent 在每个 STEP 开始前读取

---

## 当前活跃 Version

```yaml
current_version: v2
```

---

## Version 清单

| Version | 状态 | 路径 | 说明 |
|---------|:----:|------|------|
| v1 | finished | `docs/spec/versions/v1-finished/` | 首个 version，基础拓扑采集、状态引擎与 Web 监控完成交付 |
| v2 | active | `docs/spec/versions/v2/` | 当前活跃 version，历史快照回放与移动端自适应优化 |
| v3 | planned | `docs/spec/versions/v3/` | 待启动新需求迭代 |

> 完成一个 version 时：
> 1. 将目录改名 `v1` → `v1-finished`
> 2. 新建 `v2/` 目录
> 3. 把本文件 `current_version` 更新为 `v2`

---

## 目录结构

```
docs/spec/
├── README.md                     # 本文件
├── test_strategy.md              # 测试策略简报（项目级）
├── info.md                       # 测试账号与环境信息（项目级，gitignore）
├── steering/                     # 项目级恒定规则（不随 version 变更）
│   ├── project-rules.md          # 协作 / Git / 测试门槛
│   ├── tech-stack.md             # 锁定技术栈
│   └── style-guide.md            # 命名 / 错误格式 / 日志
└── versions/
    ├── v1/                       # 首个 version
    │   ├── constitution.md       # 项目宪法（vN+1 沿用 v1）
    │   ├── requirements.md       # 需求与 AC
    │   ├── design.md             # 技术设计
    │   ├── data.md               # 状态模型与配置格式（如有）
    │   ├── api.md                # 接口设计与 WebSocket 协议
    │   ├── tasks.md              # 任务清单
    │   ├── traceability.md       # 可追溯性矩阵
    │   ├── analysis-report-{{DATE}}.md
    │   ├── ac/                   # AC 验收矩阵
    │   │   └── TASK-XXX-AC.md
    │   └── reports/              # 测试报告
    │       └── REPORT_ID-report.md
    └── v2/                       # 后续 version
        └── ...
```

---

## 切换 Version 的步骤

```
1. vN 所有 TASK 完成并合并到主分支
2. 验证 vN 的 docs/spec/versions/vN/ 全部文档已定稿
3. git mv docs/spec/versions/vN docs/spec/versions/vN-finished
4. mkdir docs/spec/versions/vN+1
5. 编辑本 README.md：
   - current_version: vN+1
   - 在清单中将 vN 行标记为 finished
   - 在清单中新增 vN+1 行
6. 进入 vN+1 的 STEP 1（沿用 v1/constitution.md，无需重新生成）
```

