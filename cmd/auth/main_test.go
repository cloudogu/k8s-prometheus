package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_main(t *testing.T) {
	const (
		defaultAuthPort = 8086
		defaultSrvPort  = 8087
	)

	t.Run("should start servers", func(t *testing.T) {

		err := os.Setenv("WEB_CONFIG_FILE", "/web-config.file")
		require.NoError(t, err)
		err = os.Setenv("WEB_PRESETS_FILE", "/web-presets.file")
		require.NoError(t, err)
		err = os.Setenv("API_KEY", "myApiKey")
		require.NoError(t, err)
		err = os.Setenv("PROMETHEUS_URL", "http://prom.url")
		require.NoError(t, err)

		// Create a channel to receive signals
		sigChan := make(chan os.Signal, 1)
		// Register SIGINT to the signal channel
		signal.Notify(sigChan, syscall.SIGINT)

		go func() {
			assert.NotPanics(t, main)
		}()

		time.Sleep(100 * time.Millisecond)

		// assert correct start of the server
		var wg sync.WaitGroup

		wg.Add(2)
		go func() {
			defer wg.Done()
			assert.True(t, checkPort(defaultAuthPort, false))
		}()

		go func() {
			defer wg.Done()
			assert.True(t, checkPort(defaultSrvPort, false))
		}()

		wg.Wait()

		sendSignal(syscall.SIGINT)

		// assert graceful shutdown
		wg.Add(2)
		go func() {
			defer wg.Done()
			assert.True(t, checkPort(defaultAuthPort, true))
		}()

		go func() {
			defer wg.Done()
			assert.True(t, checkPort(defaultSrvPort, true))
		}()

		wg.Wait()
	})
}

func isPortAvailable(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		// Port is not available
		return false
	}

	defer func() {
		_ = listener.Close()
	}()

	// Port is available
	return true
}

func checkPort(port int, available bool) bool {
	ctxWithCancel, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()

	resultChan := make(chan bool)
	go func() {
	loop:
		for {
			select {
			case <-ctxWithCancel.Done():
				break loop
			default:
				if result := isPortAvailable(port); result == available {
					resultChan <- result
				}
			}

		}
	}()

	select {
	case <-ctxWithCancel.Done():
		return false
	case <-resultChan:
		return true
	}
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
