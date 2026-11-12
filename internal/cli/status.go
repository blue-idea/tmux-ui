package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/tabwriter"
	"time"

	"tmux-ui/internal/collector"
	"tmux-ui/internal/config"
	"tmux-ui/internal/detector"
	"tmux-ui/internal/model"
)

// StatusOptions CLI 实时状态查询选项
type StatusOptions struct {
	ServerURL  string // 远程或已运行服务地址，如 http://127.0.0.1:8080
	Token      string // 访问认证 Token
	JSONOutput bool   // 是否以 JSON 格式输出
	SocketName string // tmux socket 别名 (-L)
	SocketPath string // tmux socket 绝对路径 (-S)
	ConfigPath string // 智能体识别规则配置文件路径
}

// FormatStatusTable 格式化输出终端 ASCII 状态表格
func FormatStatusTable(panes []model.PaneSummary) string {
	if len(panes) == 0 {
		return "No active tmux panes detected\n"
	}

	var sb strings.Builder
	w := tabwriter.NewWriter(&sb, 0, 0, 3, ' ', 0)

	// 打印表头
	fmt.Fprintln(w, "PANE ID\tSESSION\tAGENT\tSTATE\tCOMMAND\tIDLE\tLAST PREVIEW")
	fmt.Fprintln(w, "-------\t-------\t-----\t-----\t-------\t----\t------------")

	for _, p := range panes {
		stateUpper := strings.ToUpper(string(p.State))
		stateTag := fmt.Sprintf("[%s]", stateUpper)

		agent := p.AgentType
		if agent == "" {
			agent = "-"
		}

		cmd := p.CurrentCommand
		if cmd == "" {
			cmd = "-"
		}

		idle := fmt.Sprintf("%ds", p.IdleSeconds)

		preview := "-"
		if len(p.PreviewLines) > 0 {
			last := strings.TrimSpace(p.PreviewLines[len(p.PreviewLines)-1])
			if last != "" {
				// 单行精简，防止终端折行
				last = strings.ReplaceAll(last, "\t", " ")
				if len(last) > 40 {
					last = last[:37] + "..."
				}
				preview = last
			}
		}

		sessionWin := fmt.Sprintf("%s:%s", p.SessionName, p.WindowName)
		if p.WindowName == "" {
			sessionWin = p.SessionName
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			p.PaneID,
			sessionWin,
			agent,
			stateTag,
			cmd,
			idle,
			preview,
		)
	}

	_ = w.Flush()
	return sb.String()
}

// FormatStatusJSON 将 Pane 状态列表序列化为美化的 JSON 字符串
func FormatStatusJSON(panes []model.PaneSummary) (string, error) {
	if panes == nil {
		panes = []model.PaneSummary{}
	}
	bytes, err := json.MarshalIndent(panes, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize panes to JSON: %w", err)
	}
	return string(bytes), nil
}

// FetchStatusFromServer 向正在运行的 tmux-ui HTTP 服务端请求 /api/v1/panes 状态
func FetchStatusFromServer(ctx context.Context, serverURL, token string) ([]model.PaneSummary, error) {
	serverURL = strings.TrimRight(serverURL, "/")
	reqURL := fmt.Sprintf("%s/api/v1/panes", serverURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tmux-ui server at %s: %w", serverURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server responded with status %d: %s", resp.StatusCode, string(body))
	}

	var resObj struct {
		Code int                 `json:"code"`
		Data []model.PaneSummary `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&resObj); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return resObj.Data, nil
}

// CollectLocalStatus 直接调用本地 tmux CLI 与规则推断引擎获取当前状态
func CollectLocalStatus(ctx context.Context, cfg *model.Config, socketName, socketPath string) ([]model.PaneSummary, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	tmuxColl := collector.NewTmuxCollector(collector.Options{
		CLISocketName: socketName,
		CLISocketPath: socketPath,
		Config:        cfg.Collector,
	})

	panes, err := tmuxColl.ListPanes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tmux panes: %w", err)
	}

	if len(panes) == 0 {
		return []model.PaneSummary{}, nil
	}

	procTracer := collector.NewProcessTracer(collector.ProcessTracerOptions{})
	procTree, _ := procTracer.CollectProcessTree(ctx)
	engine := detector.NewEngine()
	tracker := collector.NewPaneOutputTracker(cfg.Collector.CaptureTailLines)
	now := time.Now()

	results := make([]model.PaneSummary, 0, len(panes))
	for _, p := range panes {
		var procs []model.ProcessInfo
		if procTree != nil {
			procs = procTree.GetPaneProcesses(p.PanePID)
		}
		agentType := collector.MatchAgentType(procs, cfg.Agents)

		var matchedRule *model.AgentRule
		for i := range cfg.Agents {
			if cfg.Agents[i].Name == agentType {
				matchedRule = &cfg.Agents[i]
				break
			}
		}

		ansiOutput, _ := tmuxColl.CapturePane(ctx, p.PaneID)
		outputInfo, changed := tracker.Update(p.PaneID, p.IsDead, ansiOutput, now)

		evalResult := engine.Evaluate(detector.EvaluationInput{
			PaneID:            p.PaneID,
			IsDead:            p.IsDead,
			CurrentCommand:    p.CurrentCommand,
			AgentType:         agentType,
			TailLines:         outputInfo.TailLines,
			HasChanged:        changed,
			LastActiveAt:      outputInfo.LastActiveAt,
			Now:               now,
			InactivityTimeout: time.Duration(cfg.Collector.InactivityTimeoutS) * time.Second,
			Rule:              matchedRule,
		})

		results = append(results, model.PaneSummary{
			PaneID:         p.PaneID,
			SessionName:    p.SessionName,
			WindowID:       p.WindowID,
			WindowName:     p.WindowName,
			AgentType:      agentType,
			State:          evalResult.State,
			CurrentCommand: p.CurrentCommand,
			PanePID:        p.PanePID,
			IdleSeconds:    int64(now.Sub(outputInfo.LastActiveAt).Seconds()),
			IsDead:         p.IsDead,
			PreviewLines:   outputInfo.TailLines,
			LastActiveAt:   outputInfo.LastActiveAt,
		})
	}

	return results, nil
}

// RunStatus 执行 status 子命令的主入口
func RunStatus(ctx context.Context, opts StatusOptions, out io.Writer) error {
	var panes []model.PaneSummary
	var err error

	if opts.ServerURL != "" {
		panes, err = FetchStatusFromServer(ctx, opts.ServerURL, opts.Token)
		if err != nil {
			return err
		}
	} else {
		// 加载配置文件（如存在）
		cfgPath := opts.ConfigPath
		if cfgPath == "" {
			cfgPath = "config/agents.yaml"
		}
		cfg, _ := config.Load(cfgPath)
		panes, err = CollectLocalStatus(ctx, cfg, opts.SocketName, opts.SocketPath)
		if err != nil {
			// 若本地无运行中的 tmux 服务，返回友好空提示或空 JSON 数组
			if strings.Contains(err.Error(), "executable file not found") || strings.Contains(err.Error(), "no server running") {
				if opts.JSONOutput {
					fmt.Fprintln(out, "[]")
					return nil
				}
				fmt.Fprintln(out, "No active tmux server running locally.")
				return nil
			}
			return err
		}
	}

	if opts.JSONOutput {
		jsonStr, err := FormatStatusJSON(panes)
		if err != nil {
			return err
		}
		fmt.Fprintln(out, jsonStr)
		return nil
	}

	tableStr := FormatStatusTable(panes)
	fmt.Fprint(out, tableStr)
	return nil
}
