package proxy

import (
	"github.com/cloudogu/k8s-prometheus/auth/configuration"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CreateServer(config configuration.Configuration, av AccountValidator) *http.Server {
	r := gin.Default()

	r.Use(BasicAuth(av))
	proxyCtrl, err := NewController(config.PrometheusUrl)
	if err != nil {
		panic(err)
	}

	r.Any("/*proxyPath", proxyCtrl.Proxy)

	return &http.Server{
		Addr:    ":8086",
		Handler: r,
	}
}
