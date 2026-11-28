// 职责: 验证 Realtime URL 解析、主机规范化与请求路径拼接。
// 关联文件:
//
//	llm/internal/pkg/xurl/realtime.go
package xurl

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseRealtimeBaseURL(t *testing.T) {
	allowHost := func(*url.URL) bool { return true }
	base, err := ParseRealtimeBaseURL("WSS://例子.com:0443/v1/", "/v1", allowHost)
	require.NoError(t, err)
	require.Equal(t, "wss://xn--fsqu00a.com/v1/realtime?model=test+model", RealtimeModelURL(base, "/v1/realtime", "test model"))

	for _, raw := range []string{
		"wss://<host>", "wss://ho\"st", "wss://foo_bar.example", "wss://bad..example",
		"wss://127.1", "wss://[::1]", "wss://host:", "wss://host:0", "wss://host:99999",
		"wss://host/%76%31", "wss://host/v1//", "wss://host/v1?", "wss://host/v1#",
	} {
		_, err := ParseRealtimeBaseURL(raw, "/v1", allowHost)
		require.Error(t, err, raw)
	}
}
