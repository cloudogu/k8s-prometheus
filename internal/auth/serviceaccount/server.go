package serviceaccount

import (
	"net/http"

	"github.com/cloudogu/k8s-prometheus/tools/internal/auth/configuration"
	"github.com/gin-gonic/gin"
)

func CreateServer(config configuration.Configuration, manager manager) *http.Server {
	r := gin.Default()

	serviceAccountCtrl := NewController(manager)

	serviceAccountGroup := r.Group("/serviceaccounts")
	serviceAccountGroup.Use(ValidateAPIKey(config.ApiKey))
	serviceAccountGroup.POST("/", serviceAccountCtrl.CreateAccount)
	serviceAccountGroup.DELETE("/:consumer", serviceAccountCtrl.DeleteAccount)

	return &http.Server{
		Addr:    ":8087",
		Handler: r,
	}
}
