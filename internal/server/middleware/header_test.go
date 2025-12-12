package middleware

import (
	"net/http"
	"testing"
)

func TestExtractAPIKeyFromHeader(t *testing.T) {
	tests := []struct {
		name        string
		authHeader  string
		expectedKey string
		expectedErr string
	}{
		{
			name:        "valid bearer token",
			authHeader:  "Bearer sk-1234567890abcdef",
			expectedKey: "sk-1234567890abcdef",
			expectedErr: "",
		},
		{
			name:        "empty header",
			authHeader:  "",
			expectedKey: "",
			expectedErr: "Authorization header is required",
		},
		{
			name:        "missing Bearer prefix",
			authHeader:  "sk-1234567890abcdef",
			expectedKey: "",
			expectedErr: "Authorization header must start with 'Bearer '",
		},
		{
			name:        "Bearer with lowercase",
			authHeader:  "bearer sk-1234567890abcdef",
			expectedKey: "",
			expectedErr: "Authorization header must start with 'Bearer '",
		},
		{
			name:        "Bearer without space",
			authHeader:  "Bearersk-1234567890abcdef",
			expectedKey: "",
			expectedErr: "Authorization header must start with 'Bearer '",
		},
		{
			name:        "Bearer with empty key",
			authHeader:  "Bearer ",
			expectedKey: "",
			expectedErr: "API key is required",
		},
		{
			name:        "Bearer with only spaces",
			authHeader:  "Bearer    ",
			expectedKey: "   ",
			expectedErr: "",
		},
		{
			name:        "valid key with special characters",
			authHeader:  "Bearer sk-proj-1234567890abcdef_ghijklmnop",
			expectedKey: "sk-proj-1234567890abcdef_ghijklmnop",
			expectedErr: "",
		},
		{
			name:        "multiple Bearer prefixes",
			authHeader:  "Bearer Bearer sk-1234567890abcdef",
			expectedKey: "Bearer sk-1234567890abcdef",
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := ExtractAPIKeyFromHeader(tt.authHeader)

			if tt.expectedErr != "" {
				if err == nil {
					t.Errorf("expected error '%s', got nil", tt.expectedErr)
					return
				}

				if err.Error() != tt.expectedErr {
					t.Errorf("expected error '%s', got '%s'", tt.expectedErr, err.Error())
				}

				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if key != tt.expectedKey {
				t.Errorf("expected key '%s', got '%s'", tt.expectedKey, key)
			}
		})
	}
}

func TestExtractAPIKeyFromRequest(t *testing.T) {
	tests := []struct {
		name        string
		headers     map[string]string
		config      *APIKeyConfig
		expectedKey string
		expectedErr string
	}{
		{
			name: "Authorization header with Bearer",
			headers: map[string]string{
				"Authorization": "Bearer sk-1234567890abcdef",
			},
			config:      nil, // 使用默认配置
			expectedKey: "sk-1234567890abcdef",
			expectedErr: "",
		},
		{
			name: "X-API-Key header",
			headers: map[string]string{
				"X-API-Key": "sk-1234567890abcdef",
			},
			config:      nil,
			expectedKey: "sk-1234567890abcdef",
			expectedErr: "",
		},
		{
			name: "X-Api-Key header",
			headers: map[string]string{
				"X-Api-Key": "sk-1234567890abcdef",
			},
			config:      nil,
			expectedKey: "sk-1234567890abcdef",
			expectedErr: "",
		},
		{
			name: "API-Key header",
			headers: map[string]string{
				"API-Key": "sk-1234567890abcdef",
			},
			config:      nil,
			expectedKey: "sk-1234567890abcdef",
			expectedErr: "",
		},
		{
			name: "Authorization without Bearer prefix",
			headers: map[string]string{
				"Authorization": "sk-1234567890abcdef",
			},
			config:      nil,
			expectedKey: "sk-1234567890abcdef",
			expectedErr: "",
		},
		{
			name: "Token prefix",
			headers: map[string]string{
				"Authorization": "Token sk-1234567890abcdef",
			},
			config:      nil,
			expectedKey: "sk-1234567890abcdef",
			expectedErr: "",
		},
		{
			name: "Priority test - Authorization first",
			headers: map[string]string{
				"Authorization": "Bearer auth-key",
				"X-API-Key":     "x-api-key",
			},
			config:      nil,
			expectedKey: "auth-key",
			expectedErr: "",
		},
		{
			name: "Priority test - X-API-Key when Authorization missing",
			headers: map[string]string{
				"X-API-Key": "x-api-key",
				"API-Key":   "api-key",
			},
			config:      nil,
			expectedKey: "x-api-key",
			expectedErr: "",
		},
		{
			name: "Custom config with RequireBearer",
			headers: map[string]string{
				"Authorization": "sk-1234567890abcdef",
			},
			config: &APIKeyConfig{
				Headers:         []string{"Authorization"},
				RequireBearer:   true,
				AllowedPrefixes: []string{"Bearer "},
			},
			expectedKey: "",
			expectedErr: "Authorization header must start with 'Bearer '",
		},
		{
			name: "Custom config with custom headers",
			headers: map[string]string{
				"Custom-API-Key": "custom-key",
			},
			config: &APIKeyConfig{
				Headers:         []string{"Custom-API-Key"},
				RequireBearer:   false,
				AllowedPrefixes: []string{},
			},
			expectedKey: "custom-key",
			expectedErr: "",
		},
		{
			name: "Empty API key",
			headers: map[string]string{
				"X-API-Key": "",
			},
			config:      nil,
			expectedKey: "",
			expectedErr: "API key not found in any of the supported headers",
		},
		{
			name: "Whitespace only API key",
			headers: map[string]string{
				"X-API-Key": "   ",
			},
			config:      nil,
			expectedKey: "",
			expectedErr: "API key is required",
		},
		{
			name: "API key with leading/trailing spaces",
			headers: map[string]string{
				"X-API-Key": "  sk-1234567890abcdef  ",
			},
			config:      nil,
			expectedKey: "sk-1234567890abcdef",
			expectedErr: "",
		},
		{
			name:        "No headers provided",
			headers:     map[string]string{},
			config:      nil,
			expectedKey: "",
			expectedErr: "API key not found in any of the supported headers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 HTTP 请求
			req, err := http.NewRequest(http.MethodGet, "/test", nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			// 设置 headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			// 提取 API key
			key, err := ExtractAPIKeyFromRequest(req, tt.config)

			if tt.expectedErr != "" {
				if err == nil {
					t.Errorf("expected error '%s', got nil", tt.expectedErr)
					return
				}

				if err.Error() != tt.expectedErr {
					t.Errorf("expected error '%s', got '%s'", tt.expectedErr, err.Error())
				}

				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if key != tt.expectedKey {
				t.Errorf("expected key '%s', got '%s'", tt.expectedKey, key)
			}
		})
	}
}

func TestExtractAPIKeyFromRequestSimple(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("X-Api-Key", "simple-test-key")

	key, err := ExtractAPIKeyFromRequestSimple(req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if key != "simple-test-key" {
		t.Errorf("expected key 'simple-test-key', got '%s'", key)
	}
}

func TestDefaultAPIKeyConfig(t *testing.T) {
	config := defaultAPIKeyConfig()

	expectedHeaders := []string{"Authorization", "X-API-Key", "X-Api-Key", "API-Key", "Api-Key", "X-Goog-Api-Key", "X-Google-Api-Key"}
	if len(config.Headers) != len(expectedHeaders) {
		t.Errorf("expected %d headers, got %d", len(expectedHeaders), len(config.Headers))
	}

	for i, expected := range expectedHeaders {
		if i >= len(config.Headers) || config.Headers[i] != expected {
			t.Errorf("expected header[%d] to be '%s', got '%s'", i, expected, config.Headers[i])
		}
	}

	if config.RequireBearer {
		t.Error("expected RequireBearer to be false")
	}

	expectedPrefixes := []string{"Bearer ", "Token ", "Api-Key ", "API-Key "}
	if len(config.AllowedPrefixes) != len(expectedPrefixes) {
		t.Errorf("expected %d prefixes, got %d", len(expectedPrefixes), len(config.AllowedPrefixes))
	}
}

// BenchmarkExtractAPIKeyFromHeader 性能测试.
func BenchmarkExtractAPIKeyFromHeader(b *testing.B) {
	authHeader := "Bearer sk-1234567890abcdef"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = ExtractAPIKeyFromHeader(authHeader)
	}
}

// BenchmarkExtractAPIKeyFromRequest 性能测试.
func BenchmarkExtractAPIKeyFromRequest(b *testing.B) {
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer sk-1234567890abcdef")

	config := defaultAPIKeyConfig()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = ExtractAPIKeyFromRequest(req, config)
	}
}

// TestParseAPIKeyWithChannel 测试API密钥与渠道ID解析功能.
func TestParseAPIKeyWithChannel(t *testing.T) {
	tests := []struct {
		name              string
		rawKey            string
		expectedAPIKey    string
		expectedChannelID *int
		expectedErr       string
	}{
		{
			name:              "正常格式 - 带渠道ID",
			rawKey:            "ah-xxx#10",
			expectedAPIKey:    "ah-xxx",
			expectedChannelID: intPtr(10),
			expectedErr:       "",
		},
		{
			name:              "正常格式 - 不带渠道ID",
			rawKey:            "ah-xxx",
			expectedAPIKey:    "ah-xxx",
			expectedChannelID: nil,
			expectedErr:       "",
		},
		{
			name:              "空渠道ID - 仅有#但无ID",
			rawKey:            "ah-xxx#",
			expectedAPIKey:    "ah-xxx",
			expectedChannelID: nil,
			expectedErr:       "",
		},
		{
			name:              "空渠道ID - 仅有#和空格",
			rawKey:            "ah-xxx#  ",
			expectedAPIKey:    "ah-xxx",
			expectedChannelID: nil,
			expectedErr:       "",
		},
		{
			name:              "无效渠道ID - 非数字",
			rawKey:            "ah-xxx#abc",
			expectedAPIKey:    "",
			expectedChannelID: nil,
			expectedErr:       "invalid channel ID: abc",
		},
		{
			name:              "无效渠道ID - 包含字母和数字混合",
			rawKey:            "ah-xxx#10abc",
			expectedAPIKey:    "",
			expectedChannelID: nil,
			expectedErr:       "invalid channel ID: 10abc",
		},
		{
			name:              "无效格式 - 多个#分隔符",
			rawKey:            "ah-xxx#10#20",
			expectedAPIKey:    "",
			expectedChannelID: nil,
			expectedErr:       "invalid API key format: multiple # separators",
		},
		{
			name:              "无效格式 - 空密钥",
			rawKey:            "#10",
			expectedAPIKey:    "",
			expectedChannelID: nil,
			expectedErr:       "API key cannot be empty",
		},
		{
			name:              "无效格式 - 仅有#",
			rawKey:            "#",
			expectedAPIKey:    "",
			expectedChannelID: nil,
			expectedErr:       "API key cannot be empty",
		},
		{
			name:              "无效格式 - 空白密钥加渠道ID",
			rawKey:            "  #10",
			expectedAPIKey:    "",
			expectedChannelID: nil,
			expectedErr:       "API key cannot be empty",
		},
		{
			name:              "边界情况 - 前后有空格",
			rawKey:            "  ah-xxx#10  ",
			expectedAPIKey:    "ah-xxx",
			expectedChannelID: intPtr(10),
			expectedErr:       "",
		},
		{
			name:              "边界情况 - 渠道ID为0",
			rawKey:            "ah-xxx#0",
			expectedAPIKey:    "ah-xxx",
			expectedChannelID: intPtr(0),
			expectedErr:       "",
		},
		{
			name:              "边界情况 - 渠道ID为大数字",
			rawKey:            "ah-xxx#999999",
			expectedAPIKey:    "ah-xxx",
			expectedChannelID: intPtr(999999),
			expectedErr:       "",
		},
		{
			name:              "边界情况 - 渠道ID为负数",
			rawKey:            "ah-xxx#-10",
			expectedAPIKey:    "ah-xxx",
			expectedChannelID: intPtr(-10),
			expectedErr:       "",
		},
		{
			name:              "实际密钥格式 - OpenAI风格",
			rawKey:            "sk-proj-1234567890abcdef#5",
			expectedAPIKey:    "sk-proj-1234567890abcdef",
			expectedChannelID: intPtr(5),
			expectedErr:       "",
		},
		{
			name:              "实际密钥格式 - Anthropic风格",
			rawKey:            "sk-ant-api03-1234567890#15",
			expectedAPIKey:    "sk-ant-api03-1234567890",
			expectedChannelID: intPtr(15),
			expectedErr:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiKey, channelID, err := ParseAPIKeyWithChannel(tt.rawKey)

			// 验证错误
			if tt.expectedErr != "" {
				if err == nil {
					t.Errorf("期望错误 '%s'，但得到 nil", tt.expectedErr)
					return
				}

				if err.Error() != tt.expectedErr {
					t.Errorf("期望错误 '%s'，但得到 '%s'", tt.expectedErr, err.Error())
				}

				return
			}

			// 验证没有错误
			if err != nil {
				t.Errorf("意外的错误: %v", err)
				return
			}

			// 验证API密钥
			if apiKey != tt.expectedAPIKey {
				t.Errorf("期望API密钥 '%s'，但得到 '%s'", tt.expectedAPIKey, apiKey)
			}

			// 验证渠道ID
			if tt.expectedChannelID == nil {
				if channelID != nil {
					t.Errorf("期望渠道ID为 nil，但得到 %d", *channelID)
				}
			} else {
				if channelID == nil {
					t.Errorf("期望渠道ID为 %d，但得到 nil", *tt.expectedChannelID)
				} else if *channelID != *tt.expectedChannelID {
					t.Errorf("期望渠道ID为 %d，但得到 %d", *tt.expectedChannelID, *channelID)
				}
			}
		})
	}
}

// BenchmarkParseAPIKeyWithChannel 性能测试.
func BenchmarkParseAPIKeyWithChannel(b *testing.B) {
	rawKey := "ah-xxx#10"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _ = ParseAPIKeyWithChannel(rawKey)
	}
}

// intPtr 辅助函数，返回int指针.
func intPtr(i int) *int {
	return &i
}
