package proxy

import (
	"github.com/cloudogu/k8s-prometheus/auth/configuration"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CreateServer(config configuration.Configuration, av accountValidator) *http.Server {
	r := gin.Default()

	r.Use(BasicAuth(av))
	proxy, err := NewProxy(config.PrometheusUrl)
	if err != nil {
		panic(err)
	}

	r.Any("/*proxyPath", gin.WrapH(proxy))

	return &http.Server{
		Addr:    ":8086",
		Handler: r,
	}
}
