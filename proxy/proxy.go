package proxy

import (
	"fmt"
	"net/http/httputil"
	"net/url"
)

func NewProxy(prometheusUrl string) (*httputil.ReverseProxy, error) {
	proxyUrl, err := url.Parse(prometheusUrl)
	if err != nil {
		return nil, fmt.Errorf("error parsing proxy-url: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(proxyUrl)

	return proxy, nil
}
