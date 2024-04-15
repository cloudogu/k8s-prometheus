package proxy

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http/httputil"
	"net/url"
)

type Controller struct {
	proxy *httputil.ReverseProxy
}

func NewController(prometheusUrl string) (*Controller, error) {
	proxyUrl, err := url.Parse(prometheusUrl)
	if err != nil {
		return nil, fmt.Errorf("error parsing proxy-url: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(proxyUrl)

	return &Controller{proxy: proxy}, nil
}

func (ctrl *Controller) Proxy(c *gin.Context) {
	ctrl.proxy.ServeHTTP(c.Writer, c.Request)
}
