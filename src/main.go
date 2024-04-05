package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudogu/k8s-prometheus/auth/configuration"
	"github.com/cloudogu/k8s-prometheus/auth/prometheus"
	"github.com/cloudogu/k8s-prometheus/auth/proxy"
	"github.com/cloudogu/k8s-prometheus/auth/serviceaccount"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	config, err := configuration.ReadConfigFromEnv()
	if err != nil {
		panic(err)
	}

	manager := prometheus.NewManager(config.WebConfigFile)

	serviceAccountSrv := serviceaccount.CreateServer(config, manager)
	go func() {
		fmt.Printf("service-account-provider started on %s...\n", serviceAccountSrv.Addr)
		if err := serviceAccountSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	proxySrv := proxy.CreateServer(config, manager)
	go func() {
		fmt.Printf("auth-proxy started on %s...\n", proxySrv.Addr)
		if err := proxySrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		if err := serviceAccountSrv.Shutdown(ctx); err != nil {
			log.Fatal("Server Shutdown:", err)
		}
	}()

	go func() {
		if err := proxySrv.Shutdown(ctx); err != nil {
			log.Fatal("Server Shutdown:", err)
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("timeout of 5 seconds.")
	}
	log.Println("exiting")
}
