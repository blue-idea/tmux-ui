package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tmux-ui/internal/model"
	"tmux-ui/internal/server"
	"tmux-ui/internal/store"
)

// 创建测试用基础上下文
func setupTestServer(token string) (*server.HTTPServer, *store.MemoryStore) {
	memStore := store.NewMemoryStore()
	cfg := &model.Config{
		Agents: []model.AgentRule{
			{Name: "claude-code"},
			{Name: "aider"},
		},
	}
	s := server.NewHTTPServer(server.HTTPServerOptions{
		Store:       memStore,
		Config:      cfg,
		AuthToken:   token,
		TmuxVersion: "tmux 3.4",
	})
	return s, memStore
}

// v1.REQ-005-AC-001：GET /api/v1/health 与 GET /api/v1/config/agents
func TestHTTP_HealthAndConfig(t *testing.T) {
	srv, _ := setupTestServer("")

	t.Run("GET /api/v1/health 返回系统状态", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		rec := httptest.NewRecorder()
		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("期望状态码 200，实际为 %d", rec.Code)
		}

		var resp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Status      string `json:"status"`
				TmuxVersion string `json:"tmux_version"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("JSON 解析失败: %v", err)
		}
		if resp.Code != 0 || resp.Data.Status != "healthy" || resp.Data.TmuxVersion != "tmux 3.4" {
			t.Errorf("返回数据不符: %+v", resp)
		}
	})

	t.Run("GET /api/v1/config/agents 返回智能体规则列表", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/config/agents", nil)
		rec := httptest.NewRecorder()
		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("期望状态码 200，实际为 %d", rec.Code)
		}

		var resp struct {
			Code int `json:"code"`
			Data struct {
				Agents []model.AgentRule `json:"agents"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("JSON 解析失败: %v", err)
		}
		if len(resp.Data.Agents) != 2 {
			t.Errorf("期望返回 2 个 agent，实际为 %d", len(resp.Data.Agents))
		}
	})
}

// v1.REQ-005-AC-001：GET /api/v1/panes 列表查询与 status/session 过滤
func TestHTTP_ListPanes(t *testing.T) {
	srv, memStore := setupTestServer("")
	now := time.Now()

	memStore.UpsertPane(model.PaneDetail{
		PaneSummary: model.PaneSummary{
			PaneID:       "%1",
			SessionName:  "dev",
			State:        model.StateWorking,
			LastActiveAt: now,
		},
	}, "test")
	memStore.UpsertPane(model.PaneDetail{
		PaneSummary: model.PaneSummary{
			PaneID:       "%2",
			SessionName:  "dev",
			State:        model.StateWaitingInput,
			LastActiveAt: now,
		},
	}, "test")
	memStore.UpsertPane(model.PaneDetail{
		PaneSummary: model.PaneSummary{
			PaneID:       "%3",
			SessionName:  "prod",
			State:        model.StateIdle,
			LastActiveAt: now,
		},
	}, "test")

	t.Run("获取全部 Pane", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/panes", nil)
		rec := httptest.NewRecorder()
		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("期望状态码 200，实际为 %d", rec.Code)
		}
		var resp struct {
			Code int                 `json:"code"`
			Data []model.PaneSummary `json:"data"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		if len(resp.Data) != 3 {
			t.Errorf("期望 3 个 pane，实际为 %d", len(resp.Data))
		}
	})

	t.Run("按状态过滤 ?status=waiting_input", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/panes?status=waiting_input", nil)
		rec := httptest.NewRecorder()
		srv.Router().ServeHTTP(rec, req)

		var resp struct {
			Data []model.PaneSummary `json:"data"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		if len(resp.Data) != 1 || resp.Data[0].PaneID != "%2" {
			t.Errorf("按状态过滤错误: %+v", resp.Data)
		}
	})

	t.Run("按会话过滤 ?session=prod", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/panes?session=prod", nil)
		rec := httptest.NewRecorder()
		srv.Router().ServeHTTP(rec, req)

		var resp struct {
			Data []model.PaneSummary `json:"data"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		if len(resp.Data) != 1 || resp.Data[0].PaneID != "%3" {
			t.Errorf("按会话过滤错误: %+v", resp.Data)
		}
	})
}

// v1.REQ-005-AC-002：GET /api/v1/panes/{id} 宽容路由与详情获取
func TestHTTP_GetPaneDetail_FlexibleRouting(t *testing.T) {
	srv, memStore := setupTestServer("")
	now := time.Now()

	memStore.UpsertPane(model.PaneDetail{
		PaneSummary: model.PaneSummary{
			PaneID:       "%1",
			SessionName:  "dev",
			State:        model.StateWorking,
			LastActiveAt: now,
		},
		RawOutputANSI: "\x1b[32mhello terminal\x1b[0m",
	}, "test")

	t.Run("支持裸数字 ID /api/v1/panes/1", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/panes/1", nil)
		rec := httptest.NewRecorder()
		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("裸数字 ID 期望 200，实际为 %d: %s", rec.Code, rec.Body.String())
		}
		var resp struct {
			Code int              `json:"code"`
			Data model.PaneDetail `json:"data"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp.Data.PaneID != "%1" || resp.Data.RawOutputANSI != "\x1b[32mhello terminal\x1b[0m" {
			t.Errorf("详情数据返回不符: %+v", resp.Data)
		}
	})

	t.Run("支持编码字符 /api/v1/panes/%251", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/panes/%251", nil)
		rec := httptest.NewRecorder()
		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("URL 编码 ID 期望 200，实际为 %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("不存在的 Pane 返回 404 及标准错误 JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/panes/99", nil)
		rec := httptest.NewRecorder()
		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("不存在的 Pane 期望 404，实际为 %d", rec.Code)
		}

		var errResp struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
		if errResp.Error.Code != "PANE_NOT_FOUND" {
			t.Errorf("错误码期望 PANE_NOT_FOUND，实际为 %s", errResp.Error.Code)
		}
	})
}

// v1.REQ-005-AC-004：Token 认证中间件
func TestHTTP_AuthMiddleware(t *testing.T) {
	const validToken = "test-secret-token-123456"
	srv, _ := setupTestServer(validToken)

	t.Run("未携带 Token 返回 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/panes", nil)
		rec := httptest.NewRecorder()
		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("未携带 Token 期望 401，实际为 %d", rec.Code)
		}

		var errResp struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
		if errResp.Error.Code != "UNAUTHORIZED" {
			t.Errorf("期望错误码 UNAUTHORIZED，实际为 %s", errResp.Error.Code)
		}
	})

	t.Run("携带无效 Token 返回 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/panes", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec := httptest.NewRecorder()
		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("无效 Token 期望 401，实际为 %d", rec.Code)
		}
	})

	t.Run("携带有效 Header Token 返回 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/panes", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()
		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("有效 Token 期望 200，实际为 %d", rec.Code)
		}
	})

	t.Run("携带有效 Query Param Token 返回 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/panes?token="+validToken, nil)
		rec := httptest.NewRecorder()
		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("有效 Query Token 期望 200，实际为 %d", rec.Code)
		}
	})
}
