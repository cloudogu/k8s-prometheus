package proxy

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Controller struct {
	prometheusUrl *url.URL
}

func NewController(prometheusUrl string) (*Controller, error) {
	proxyUrl, err := url.Parse(prometheusUrl)
	if err != nil {
		return nil, fmt.Errorf("error parsing proxy-url: %w", err)
	}

	return &Controller{prometheusUrl: proxyUrl}, nil
}

func (ctrl *Controller) Proxy(c *gin.Context) {
	proxy := httputil.NewSingleHostReverseProxy(ctrl.prometheusUrl)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = ctrl.prometheusUrl.Host
		req.URL.Scheme = ctrl.prometheusUrl.Scheme
		req.URL.Host = ctrl.prometheusUrl.Host
		req.URL.Path = c.Param("proxyPath")
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}
