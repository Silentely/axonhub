// 职责: 校验并构造地域和 Workspace 专属 Qwen 实时地址。
// 关联文件:
//
//	llm/realtime/capability.go
//	internal/server/biz/channel_endpoint.go
package realtime

import (
	"errors"
	"net/url"
	"regexp"

	"github.com/looplj/axonhub/llm/internal/pkg/xurl"
)

const qwenRealtimePrefix = "/api-ws/v1"
const qwenRealtimePath = qwenRealtimePrefix + "/realtime"

var trustedHost = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.(?:cn-beijing|ap-southeast-1)\.maas\.aliyuncs\.com$`)

func ValidateEndpoint(baseURL, path string) error {
	_, err := parseEndpoint(baseURL, path)
	return err
}

func parseEndpoint(baseURL, path string) (*url.URL, error) {
	u, err := xurl.ParseRealtimeBaseURL(baseURL, qwenRealtimePrefix, func(u *url.URL) bool {
		return trustedHost.MatchString(u.Host)
	})
	if err != nil {
		return nil, errors.New("Qwen realtime requires a trusted Workspace regional WSS host with an optional /api-ws/v1 path")
	}
	if path != qwenRealtimePath {
		return nil, errors.New("Qwen realtime requires /api-ws/v1/realtime")
	}
	return u, nil
}
func EndpointURL(baseURL, path, model string) (string, error) {
	base, err := parseEndpoint(baseURL, path)
	if err != nil {
		return "", err
	}
	if !IsAssociatedModel(model) {
		return "", errUnsupportedModel(model)
	}
	return xurl.RealtimeModelURL(base, qwenRealtimePath, model), nil
}
