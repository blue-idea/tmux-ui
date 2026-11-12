package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"tmux-ui/internal/model"
	"tmux-ui/internal/store"
)

// SnapshotReader 抽象快照只读查询接口
type SnapshotReader interface {
	ListSnapshots(paneID string, since int64, limit int) ([]model.SnapshotMetadata, error)
	GetSnapshot(paneID string, snapshotID string) (*model.SnapshotRecord, error)
}

// HTTPServerOptions 初始化 HTTPServer 的选项
type HTTPServerOptions struct {
	Store          *store.MemoryStore
	Config         *model.Config
	AuthToken      string
	TmuxVersion    string
	WSHub          *WSHub
	StaticFS       fs.FS
	HistoryStorage SnapshotReader
}

// HTTPServer 提供 REST API 服务与路由分发
type HTTPServer struct {
	router         chi.Router
	store          *store.MemoryStore
	config         *model.Config
	authToken      string
	tmuxVersion    string
	startTime      time.Time
	wsHub          *WSHub
	staticFS       fs.FS
	historyStorage SnapshotReader
}

// GetClientIP 提取请求的真实客户端公网 IP，优先识别 Cloudflare CF-Connecting-IP 与反向代理头
func GetClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}

	// 1. Cloudflare 边缘真实用户 IP (CF-Connecting-IP / True-Client-IP)
	if cfIP := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); cfIP != "" {
		return cfIP
	}
	if trueIP := strings.TrimSpace(r.Header.Get("True-Client-IP")); trueIP != "" {
		return trueIP
	}

	// 2. X-Forwarded-For 链中的第一个真实 IP
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			first := strings.TrimSpace(parts[0])
			if first != "" {
				return first
			}
		}
	}

	// 3. X-Real-IP
	if xRealIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); xRealIP != "" {
		return xRealIP
	}

	// 4. 直连 RemoteAddr 解析 Host
	remote := strings.TrimSpace(r.RemoteAddr)
	if remote != "" {
		if host, _, err := net.SplitHostPort(remote); err == nil && host != "" {
			return host
		}
		return remote
	}

	return ""
}

// CloudflareRealIPMiddleware 将 Cloudflare 与反向代理客户端 IP 规范化写入 RemoteAddr
func CloudflareRealIPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if clientIP := GetClientIP(r); clientIP != "" {
			_, port, err := net.SplitHostPort(r.RemoteAddr)
			if err == nil && port != "" {
				r.RemoteAddr = net.JoinHostPort(clientIP, port)
			} else {
				r.RemoteAddr = clientIP
			}
		}
		next.ServeHTTP(w, r)
	})
}

// NewHTTPServer 构造并初始化路由规则
func NewHTTPServer(opts HTTPServerOptions) *HTTPServer {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(CloudflareRealIPMiddleware)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	hub := opts.WSHub
	if hub == nil {
		hub = NewWSHub(opts.Store, opts.AuthToken)
		go hub.Run()
	}

	s := &HTTPServer{
		router:         r,
		store:          opts.Store,
		config:         opts.Config,
		authToken:      opts.AuthToken,
		tmuxVersion:    opts.TmuxVersion,
		startTime:      time.Now(),
		wsHub:          hub,
		staticFS:       opts.StaticFS,
		historyStorage: opts.HistoryStorage,
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
			pr.Get("/panes/{id}/snapshots", s.handleListSnapshots)
			pr.Get("/panes/{id}/snapshots/{snapshot_id}", s.handleGetSnapshotDetail)
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

// normalizePaneID 对 Pane ID 进行 URL 解码并规整为带 % 前缀的标准格式
func normalizePaneID(rawID string) string {
	unescaped, err := url.PathUnescape(rawID)
	if err == nil {
		rawID = unescaped
	}
	if !strings.HasPrefix(rawID, "%") {
		rawID = "%" + rawID
	}
	return rawID
}

// handleGetPaneDetail 获取单个 Pane 详情（支持宽容路由：裸数字 1 或 URL 编码 %251）
func (s *HTTPServer) handleGetPaneDetail(w http.ResponseWriter, r *http.Request) {
	paneID := normalizePaneID(chi.URLParam(r, "id"))

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

// handleListSnapshots 查询指定 Pane 的历史快照元数据列表
func (s *HTTPServer) handleListSnapshots(w http.ResponseWriter, r *http.Request) {
	paneID := normalizePaneID(chi.URLParam(r, "id"))
	if s.historyStorage == nil {
		renderSuccess(w, http.StatusOK, []model.SnapshotMetadata{})
		return
	}

	var since int64 = 0
	if sinceStr := r.URL.Query().Get("since"); sinceStr != "" {
		if val, err := strconv.ParseInt(sinceStr, 10, 64); err == nil {
			since = val
		}
	}

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			limit = val
		}
	}

	list, err := s.historyStorage.ListSnapshots(paneID, since, limit)
	if err != nil {
		renderError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve snapshots", nil)
		return
	}

	renderSuccess(w, http.StatusOK, list)
}

// handleGetSnapshotDetail 获取指定 Pane 的单条历史快照完整详情（含 ANSI 内容）
func (s *HTTPServer) handleGetSnapshotDetail(w http.ResponseWriter, r *http.Request) {
	paneID := normalizePaneID(chi.URLParam(r, "id"))
	snapshotID := chi.URLParam(r, "snapshot_id")

	if s.historyStorage == nil {
		renderError(w, http.StatusNotFound, "SNAPSHOT_NOT_FOUND", fmt.Sprintf("Snapshot '%s' not found for pane '%s'", snapshotID, paneID), nil)
		return
	}

	record, err := s.historyStorage.GetSnapshot(paneID, snapshotID)
	if err != nil || record == nil {
		renderError(w, http.StatusNotFound, "SNAPSHOT_NOT_FOUND", fmt.Sprintf("Snapshot '%s' not found for pane '%s'", snapshotID, paneID), nil)
		return
	}

	renderSuccess(w, http.StatusOK, record)
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
