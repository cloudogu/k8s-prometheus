package main

import (
	"fmt"
	"github.com/cloudogu/k8s-prometheus/serviceaccount/configuration"
	"github.com/cloudogu/k8s-prometheus/serviceaccount/prometheus"
	"github.com/cloudogu/k8s-prometheus/serviceaccount/serviceaccount"
	"github.com/gin-gonic/gin"
)

func main() {
	config, err := configuration.ReadConfigFromEnv()
	if err != nil {
		panic(err)
	}

	manager := prometheus.NewManager(config.WebConfigFile)
	controller := serviceaccount.NewController(manager)

	r := gin.Default()
	r.Use(serviceaccount.ValidateAPIKey(config.ApiKey))
	r.POST("/serviceaccounts", controller.CreateAccount)
	r.DELETE("/serviceaccounts/:consumer", controller.DeleteAccount)

	fmt.Println("service-account-sidecar started on port 8080...")

	// listen and serve on 0.0.0.0:8080
	if err := r.Run(":8087"); err != nil {
		panic(err)
	}
}
