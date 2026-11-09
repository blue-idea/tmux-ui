# tmux 多智能体可视化工具

## 1. 核心思路

不需要侵入 tmux 或 agent，只靠 tmux 自带的命令行接口就能拿到几乎所有信息，这也是 mac/Linux 都能兼容的基础。

| 需求                        | tmux 能力                                                    |
| --------------------------- | ------------------------------------------------------------ |
| 拓扑（session/window/pane） | `tmux list-panes -a -F '<format>'`                           |
| 实时输出预览                | `tmux capture-pane -p -e -t %id`（`-e` 保留 ANSI 颜色）      |
| 事件推送（增删 pane、输出） | Control Mode：`tmux -C attach`，输出 `%output`、`%window-add` 等事件 |
| 活跃度                      | `#{pane_activity}`、`#{window_activity}`、`#{pane_dead}`     |
| 当前进程                    | `#{pane_current_command}`、`#{pane_pid}` 再配合 `ps` 查进程树 |

MVP 阶段用轮询（500ms~1s）`list-panes` 和 `capture-pane`，最简单稳定。之后再升级到 Control Mode 做低延迟推送。

## 2. 架构

```
Browser (React/Vue + xterm.js)
        │  WebSocket + REST
Backend (本机常驻进程)
 ├─ Collector：调用 tmux CLI，采集拓扑和 pane 内容
 ├─ Detector：判断每个 agent 的状态
 ├─ State Store：内存快照 + 短期历史（状态时间线）
 └─ API/WS：推送 diff
        │
      tmux server (本机 socket)
```

## 3. 技术选型

- **后端：Go（推荐）**。能编译成单个二进制并用 `embed` 内嵌前端，mac（arm64/amd64）和 Linux 直接分发，用户无需装运行时。备选 Node.js（Fastify + ws），开发更快，但分发要装 Node。
- **前端：React + Vite + Tailwind**，终端渲染用 **xterm.js**，可以直接渲染带 ANSI 色彩的 `capture-pane` 内容。
- **配置：** YAML，用来定义各类 agent 的识别规则。

## 4. 难点：怎么判断 agent 状态

这是整个工具的价值所在，建议用「规则引擎 + 启发式」：

**状态模型：** `working` / `idle` / `waiting_input`（等待权限确认或提问）/ `error` / `exited`

**检测信号（按优先级）：**

1. `pane_dead` 为 1，状态为 `exited`
2. 匹配 pane 末尾 N 行的正则，比如：
   - Claude Code：出现 `esc to interrupt` 表示 working；出现 `Do you want to proceed?` 表示 waiting_input
   - Codex、Gemini CLI、Aider 等各有自己的提示符和 spinner 特征
3. 输出内容哈希在 N 秒内是否变化：有变化是 working，长时间无变化是 idle
4. 兜底：`pane_current_command` 是 shell（zsh/bash），说明 agent 已退出

规则放在 YAML 里可扩展：

```yaml
agents:
  - name: claude-code
    match_process: ["claude", "node"]
    states:
      working:       { tail_regex: "esc to interrupt" }
      waiting_input: { tail_regex: "(Do you want to|\\(y/n\\))" }
```

## 5. 界面设想

- **总览墙：** 每个 pane 一张卡片，显示 session/window 名、agent 类型、状态灯、最近几行输出、空闲时长。
- **详情页：** xterm.js 实时展示单个 pane，并带状态变化时间线。
- **筛选与通知：** 按状态筛选；出现 `waiting_input` 或 `error` 时给浏览器通知，这对多 agent 并行场景最实用。
- **可选（二期）：** 从网页给 pane `send-keys`，例如一键回复 y。

## 6. 兼容性与安全注意

- 各 tmux 版本的 format 变量略有差异，启动时用 `tmux -V` 检测并降级。
- `ps` 参数在 macOS（BSD）和 Linux（procps）上不同，用 `ps -axo pid=,ppid=,command=` 这类交集写法。
- 支持 `-L socket名` / `-S 路径`，以便监控多个 tmux server。
- 默认只监听 `127.0.0.1`，加随机 token。`send-keys` 等于远程执行命令，默认只读，写操作需显式开启。
- 不同机器上的 tmux 可以通过 SSH 隧道或在远端也跑一个 agent 进程来汇聚，放到后期做。