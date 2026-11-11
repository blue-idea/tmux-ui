package server

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"tmux-ui/internal/model"
	"tmux-ui/internal/store"
)

// HTTPServerOptions 初始化 HTTPServer 的选项
type HTTPServerOptions struct {
	Store       *store.MemoryStore
	Config      *model.Config
	AuthToken   string
	TmuxVersion string
	WSHub       *WSHub
	StaticFS    fs.FS
}

// HTTPServer 提供 REST API 服务与路由分发
type HTTPServer struct {
	router      chi.Router
	store       *store.MemoryStore
	config      *model.Config
	authToken   string
	tmuxVersion string
	startTime   time.Time
	wsHub       *WSHub
	staticFS    fs.FS
}

// NewHTTPServer 构造并初始化路由规则
func NewHTTPServer(opts HTTPServerOptions) *HTTPServer {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	hub := opts.WSHub
	if hub == nil {
		hub = NewWSHub(opts.Store, opts.AuthToken)
		go hub.Run()
	}

	s := &HTTPServer{
		router:      r,
		store:       opts.Store,
		config:      opts.Config,
		authToken:   opts.AuthToken,
		tmuxVersion: opts.TmuxVersion,
		startTime:   time.Now(),
		wsHub:       hub,
		staticFS:    opts.StaticFS,
	}

	s.setupRoutes()
	return s
}

// Router 返回内部 chi.Router 实例
func (s *HTTPServer) Router() chi.Router {
	return s.router
}

// WSHub 返回内部 WebSocket Hub 实例
func (s *HTTPServer) WSHub() *WSHub {
	return s.wsHub
}

func (s *HTTPServer) setupRoutes() {
	// WebSocket 实时全双工连接
	s.router.Get("/ws", s.wsHub.HandleWS)

	s.router.Route("/api/v1", func(r chi.Router) {
		// 健康检查无需 Token 鉴权
		r.Get("/health", s.handleHealth)

		// 其余受保护路由组
		r.Group(func(pr chi.Router) {
			if s.authToken != "" {
				pr.Use(s.authMiddleware)
			}
			pr.Get("/config/agents", s.handleConfigAgents)
			pr.Get("/panes", s.handleListPanes)
			pr.Get("/panes/{id}", s.handleGetPaneDetail)
		})
	})

	// 静态资源与前端 SPA 路由回退
	if s.staticFS != nil {
		fileServer := http.FileServer(http.FS(s.staticFS))
		s.router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/")
			if path == "" {
				fileServer.ServeHTTP(w, r)
				return
			}
			if f, err := s.staticFS.Open(path); err == nil {
				_ = f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
			indexData, err := fs.ReadFile(s.staticFS, "index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(indexData)
		})
	}
}

// authMiddleware 校验 Bearer Token 或 URL Token
func (s *HTTPServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ""

		// 1. 尝试从 Authorization Header 中提取
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// 2. 尝试从 URL Query 中提取
		if token == "" {
			token = r.URL.Query().Get("token")
		}

		if token != s.authToken {
			renderError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication token is missing or invalid", nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleHealth 处理系统健康检查
func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	activeCount := 0
	if s.store != nil {
		activeCount = len(s.store.ListPanes())
	}
	uptime := int64(time.Since(s.startTime).Seconds())

	renderSuccess(w, http.StatusOK, map[string]interface{}{
		"status":         "healthy",
		"tmux_version":   s.tmuxVersion,
		"active_panes":   activeCount,
		"uptime_seconds": uptime,
	})
}

// handleConfigAgents 获取当前智能体识别配置
func (s *HTTPServer) handleConfigAgents(w http.ResponseWriter, r *http.Request) {
	agents := []model.AgentRule{}
	if s.config != nil {
		agents = s.config.Agents
	}
	renderSuccess(w, http.StatusOK, map[string]interface{}{
		"agents": agents,
	})
}

// handleListPanes 获取所有 Pane 摘要列表（支持 status 和 session 过滤）
func (s *HTTPServer) handleListPanes(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		renderSuccess(w, http.StatusOK, []model.PaneSummary{})
		return
	}

	panes := s.store.ListPanes()
	statusFilter := r.URL.Query().Get("status")
	sessionFilter := r.URL.Query().Get("session")

	filtered := make([]model.PaneSummary, 0, len(panes))
	for _, p := range panes {
		if statusFilter != "" && string(p.State) != statusFilter {
			continue
		}
		if sessionFilter != "" && p.SessionName != sessionFilter {
			continue
		}
		filtered = append(filtered, p)
	}

	renderSuccess(w, http.StatusOK, filtered)
}

// handleGetPaneDetail 获取单个 Pane 详情（支持宽容路由：裸数字 1 或 URL 编码 %251）
func (s *HTTPServer) handleGetPaneDetail(w http.ResponseWriter, r *http.Request) {
	rawID := chi.URLParam(r, "id")
	// URL 解码
	unescaped, err := url.PathUnescape(rawID)
	if err == nil {
		rawID = unescaped
	}

	// 宽容解析：裸数字自动补充 % 前缀
	paneID := rawID
	if !strings.HasPrefix(paneID, "%") {
		paneID = "%" + paneID
	}

	if s.store == nil {
		renderError(w, http.StatusNotFound, "PANE_NOT_FOUND", "Target tmux pane does not exist", map[string]string{"pane_id": paneID})
		return
	}

	detail, exists := s.store.GetPane(paneID)
	if !exists {
		renderError(w, http.StatusNotFound, "PANE_NOT_FOUND", "Target tmux pane does not exist", map[string]string{"pane_id": paneID})
		return
	}

	renderSuccess(w, http.StatusOK, detail)
}

// renderSuccess 返回统一成功响应格式
func renderSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

// renderError 返回统一错误响应格式
func renderError(w http.ResponseWriter, status int, code, message string, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	errObj := map[string]interface{}{
		"code":    code,
		"message": message,
	}
	if details != nil {
		errObj["details"] = details
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": errObj,
	})
}
