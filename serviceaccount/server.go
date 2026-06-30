package serviceaccount

import (
	"net/http"

	"github.com/cloudogu/k8s-prometheus/auth/configuration"
	"github.com/gin-gonic/gin"
)

func CreateServer(config configuration.Configuration, manager manager) *http.Server {
	r := gin.Default()

	serviceAccountCtrl := NewController(manager)

	serviceAccountGroup := r.Group("/serviceaccounts")
	serviceAccountGroup.Use(ValidateAPIKey(config.ApiKey))
	serviceAccountGroup.PUT("/", serviceAccountCtrl.CreateOrUpdateAccount)
	serviceAccountGroup.DELETE("/:consumer", serviceAccountCtrl.DeleteAccount)
	serviceAccountGroup.HEAD("/:consumer", serviceAccountCtrl.GetAccount)
	serviceAccountGroup.GET("/:consumer", serviceAccountCtrl.GetAccount)

	return &http.Server{
		Addr:    ":8087",
		Handler: r,
	}
}
