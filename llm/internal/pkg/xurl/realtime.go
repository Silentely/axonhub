// 职责: 校验 Realtime Base URL 并构造模型连接地址。
// 关联文件:
//
//	llm/realtime/endpoint.go
//	llm/transformer/openai/realtime/native.go
//	llm/transformer/openai/realtime/translation.go
package xurl

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

func ParseRealtimeBaseURL(baseURL, prefix string, validHost func(*url.URL) bool) (*url.URL, error) {
	u, err := url.Parse(baseURL)
	if err != nil || u.Scheme != "wss" || u.Hostname() == "" ||
		u.User != nil || u.RawPath != "" || strings.ContainsAny(baseURL, "?#") || validHost == nil || !validHost(u) {
		return nil, errors.New("invalid Realtime WSS base URL")
	}
	if strings.HasSuffix(u.Host, ":") {
		return nil, errors.New("empty Realtime port")
	}
	if port := u.Port(); port != "" {
		value, err := strconv.ParseUint(port, 10, 16)
		if err != nil || value == 0 {
			return nil, errors.New("invalid Realtime port")
		}
	}
	if path := strings.TrimSuffix(u.Path, "/"); path != "" && path != prefix {
		return nil, errors.New("invalid Realtime base URL path")
	}
	return u, nil
}

func RealtimeModelURL(base *url.URL, path, model string) string {
	u := *base
	u.Path = path
	u.RawQuery = url.Values{"model": {model}}.Encode()
	return u.String()
}
