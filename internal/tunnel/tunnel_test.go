package tunnel_test

import (
	"testing"
	"tmux-ui/internal/tunnel"
)

// v3.REQ-001-AC-003: 密码学安全随机 Token 生成
func TestGenerateSecureToken(t *testing.T) {
	token1, err := tunnel.GenerateSecureToken()
	if err != nil {
		t.Fatalf("生成随机 Token 失败: %v", err)
	}
	if len(token1) < 32 {
		t.Errorf("期望 Token 长度 >= 32 字符，实际长度为 %d: %s", len(token1), token1)
	}

	token2, err := tunnel.GenerateSecureToken()
	if err != nil {
		t.Fatalf("二次生成随机 Token 失败: %v", err)
	}
	if token1 == token2 {
		t.Errorf("连续生成的两次随机 Token 不应相同: %s vs %s", token1, token2)
	}
}

// v3.REQ-001-AC-001: 公网免密 URL 格式化与 Query 拼接
func TestFormatPublicURL(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		token     string
		expected  string
	}{
		{
			name:     "基础 HTTPS 域名带斜杠",
			baseURL:  "https://tmux.example.com/",
			token:    "sec123",
			expected: "https://tmux.example.com/?token=sec123",
		},
		{
			name:     "基础域名无末尾斜杠",
			baseURL:  "https://tmux.example.com",
			token:    "sec123",
			expected: "https://tmux.example.com/?token=sec123",
		},
		{
			name:     "已有 Query 参数保留拼接",
			baseURL:  "https://tmux.example.com/?view=overview",
			token:    "sec123",
			expected: "https://tmux.example.com/?token=sec123&view=overview",
		},
		{
			name:     "无 Token 时保留原始 URL",
			baseURL:  "https://tmux.example.com",
			token:    "",
			expected: "https://tmux.example.com",
		},
		{
			name:     "URL 包含特殊字符转义编码",
			baseURL:  "https://tmux.example.com",
			token:    "token+with special=chars",
			expected: "https://tmux.example.com/?token=token%2Bwith+special%3Dchars",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tunnel.FormatPublicURL(tc.baseURL, tc.token)
			if actual != tc.expected {
				t.Errorf("FormatPublicURL(%q, %q) = %q, 期望 %q", tc.baseURL, tc.token, actual, tc.expected)
			}
		})
	}
}

// v3.REQ-001-AC-003: 公网模式强制鉴权防裸奔保障
func TestEnsureTunnelSecurity(t *testing.T) {
	t.Run("开启隧道且未配置 Token 时自动生成", func(t *testing.T) {
		token, generated, err := tunnel.EnsureTunnelSecurity(true, "")
		if err != nil {
			t.Fatalf("EnsureTunnelSecurity 出错: %v", err)
		}
		if !generated {
			t.Errorf("期望标记 generated 为 true")
		}
		if len(token) < 32 {
			t.Errorf("自动生成的 Token 长度不足 32: %s", token)
		}
	})

	t.Run("开启隧道且已有用户配置 Token 时保持原样", func(t *testing.T) {
		token, generated, err := tunnel.EnsureTunnelSecurity(true, "user_custom_token")
		if err != nil {
			t.Fatalf("EnsureTunnelSecurity 出错: %v", err)
		}
		if generated {
			t.Errorf("期望标记 generated 为 false")
		}
		if token != "user_custom_token" {
			t.Errorf("期望保持用户自定义 Token，实际为: %s", token)
		}
	})

	t.Run("未开启隧道且未配置 Token 时不自动生成", func(t *testing.T) {
		token, generated, err := tunnel.EnsureTunnelSecurity(false, "")
		if err != nil {
			t.Fatalf("EnsureTunnelSecurity 出错: %v", err)
		}
		if generated {
			t.Errorf("未开启隧道时 generated 应当为 false")
		}
		if token != "" {
			t.Errorf("未开启隧道时 token 应当为空，实际为: %s", token)
		}
	})
}
