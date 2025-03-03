package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudogu/k8s-prometheus/auth/configuration"
	"github.com/cloudogu/k8s-prometheus/auth/prometheus"
	"github.com/cloudogu/k8s-prometheus/auth/proxy"
	"github.com/cloudogu/k8s-prometheus/auth/serviceaccount"

	"github.com/gin-gonic/gin"
)

func main() {
	config, err := configuration.ReadConfigFromEnv()
	if err != nil {
		panic(err)
	}

	gin.SetMode(gin.ReleaseMode)

	configureLogger(config)

	webPresets := &prometheus.WebConfig{}
	if config.WebPresetsFile != "" {
		// if the file does not exist, an empty object will be returned
		webPresets, err = prometheus.NewWebConfigFileReaderWriter(config.WebPresetsFile).ReadWebConfig()
		if err != nil {
			panic(fmt.Errorf("failed to parse web presets file: %w", err))
		}
	}

	manager := prometheus.NewManager(config.WebConfigFile, webPresets)

	serviceAccountSrv := serviceaccount.CreateServer(config, manager)
	go func() {
		slog.Info("service-account-provider started", "addr", serviceAccountSrv.Addr)
		if err := serviceAccountSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("error starting service-account server.", "err", err)
		}
	}()

	proxySrv := proxy.CreateServer(config, manager)
	go func() {
		slog.Info("auth-proxy started", "addr", proxySrv.Addr)
		if err := proxySrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("error starting auth-proxy server.", "err", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		slog.Info("stopping service-account server")
		if err := serviceAccountSrv.Shutdown(ctx); err != nil {
			slog.Error("error stopping service-account server.", "err", err)
		}
	}()

	go func() {
		slog.Info("stopping auth-proxy server")
		if err := proxySrv.Shutdown(ctx); err != nil {
			slog.Error("error stopping auth-proxy server.", "err", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutdown-timeout of 5 seconds reached")
	slog.Info("exiting")
}

func configureLogger(conf configuration.Configuration) {
	var level slog.Level
	var err = level.UnmarshalText([]byte(conf.LogLevel))
	if err != nil {
		slog.Error("error parsing log level. Setting log level to INFO.", "err", err)
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: false,
		Level:     level,
	}))
	slog.SetDefault(logger)

	slog.Info("configured logger", "level", level.String())
}
