package main

import (
	"context"
	"github.com/cloudogu/k8s-prometheus/auth/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func Test_main(t *testing.T) {
	t.Run("should start servers", func(t *testing.T) {

		err := os.Setenv("WEB_CONFIG_FILE", "/web-config.file")
		require.NoError(t, err)
		err = os.Setenv("API_KEY", "myApiKey")
		require.NoError(t, err)
		err = os.Setenv("PROMETHEUS_URL", "http://prom.url")
		require.NoError(t, err)

		// Create a channel to receive signals
		sigChan := make(chan os.Signal, 1)
		// Register SIGINT to the signal channel
		signal.Notify(sigChan, syscall.SIGINT)

		// Create a timer for 5 seconds
		timer := time.NewTimer(5 * time.Second)

		assert.NotPanics(t, func() {
			go main()
		})

		// Wait for either the timer to expire or SIGINT to be received
		select {
		case <-timer.C:
			// Timer expired, send SIGINT signal to the process
			sendSignal(syscall.SIGINT)
		case <-sigChan:
			// SIGINT received, do nothing (the program will exit gracefully)
		}
	})
}

func Test_configureLogger(t *testing.T) {
	t.Run("should configure logger with log-level from config", func(t *testing.T) {
		conf := configuration.Configuration{LogLevel: "DEBUG"}

		configureLogger(conf)

		textHandler, ok := slog.Default().Handler().(*slog.TextHandler)
		require.True(t, ok)
		assert.True(t, textHandler.Enabled(context.TODO(), slog.LevelDebug))
	})

	t.Run("should configure logger with log-level info if config not valid", func(t *testing.T) {
		conf := configuration.Configuration{LogLevel: "NO_NO_LOG"}

		configureLogger(conf)

		textHandler, ok := slog.Default().Handler().(*slog.TextHandler)
		require.True(t, ok)
		assert.True(t, textHandler.Enabled(context.TODO(), slog.LevelInfo))
	})
}

// Function to send signal to the process
func sendSignal(sig os.Signal) {
	pid := os.Getpid()
	process, err := os.FindProcess(pid)
	if err != nil {
		// Handle error
		return
	}
	process.Signal(sig)
}
