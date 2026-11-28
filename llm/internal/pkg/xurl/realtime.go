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
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/idna"
)

var numericHostSuffix = regexp.MustCompile(`(^|\.)([0-9]+|0x[0-9a-f]*)\.?$`)

var realtimeDNSLabel = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

func ParseRealtimeBaseURL(baseURL, prefix string, validHost func(*url.URL) bool) (*url.URL, error) {
	u, err := url.Parse(baseURL)
	if err != nil || u.Scheme != "wss" || u.Hostname() == "" ||
		u.User != nil || strings.ContainsAny(baseURL, "?#") {
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
	if path := strings.TrimSuffix(u.Path, "/"); u.RawPath != "" || (path != "" && path != prefix) {
		return nil, errors.New("invalid Realtime base URL path")
	}
	host, err := idna.Lookup.ToASCII(u.Hostname())
	if err != nil || validHost == nil {
		return nil, errors.New("invalid Realtime hostname")
	}
	for _, label := range strings.Split(strings.TrimSuffix(host, "."), ".") {
		if !realtimeDNSLabel.MatchString(label) {
			return nil, errors.New("invalid Realtime hostname")
		}
	}
	// 供应商仍可按原始地址施加规则，避免规范化放宽百炼域名白名单。
	if numericHostSuffix.MatchString(host) || !validHost(u) {
		return nil, errors.New("unsupported Realtime hostname")
	}
	if port := strings.TrimLeft(u.Port(), "0"); port != "" && port != "443" {
		host += ":" + port
	}
	u.Host = host
	return u, nil
}

func RealtimeModelURL(base *url.URL, path, model string) string {
	u := *base
	u.Path = path
	u.RawQuery = url.Values{"model": {model}}.Encode()
	return u.String()
}
