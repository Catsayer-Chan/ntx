// Package ping_test provides tests for the ping package
package ping

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/catsayer/ntx/pkg/errors"
	"github.com/catsayer/ntx/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTCPServer sets up a simple TCP server on a random port for testing.
// It returns the listener and the address string.
func setupTCPServer(t *testing.T) (*net.TCPListener, string) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	require.NoError(t, err, "Failed to resolve TCP address")

	listener, err := net.ListenTCP("tcp", addr)
	require.NoError(t, err, "Failed to listen on TCP address")

	// Goroutine to accept connections so the test doesn't block
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return // Listener was closed
			}
			conn.Close()
		}
	}()

	return listener, listener.Addr().String()
}

func TestTCPPinger_Ping(t *testing.T) {
	// Case 1: Successful ping to a live server
	t.Run("Success", func(t *testing.T) {
		server, addr := setupTCPServer(t)
		defer server.Close()

		pinger := NewTCPPinger()
		defer pinger.Close()

		opts := &types.PingOptions{
			Count:   3,
			Timeout: time.Second,
		}

		ctx := context.Background()
		result, err := pinger.Ping(ctx, addr, opts)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 3, result.Statistics.Sent)
		assert.Equal(t, 3, result.Statistics.Received)
		assert.Equal(t, 0.0, result.Statistics.LossRate)
		assert.NotZero(t, result.Statistics.AvgRTT)
		assert.Equal(t, types.StatusSuccess, result.Status)
	})

	// Case 2: Ping to a non-existent server (timeout)
	t.Run("Timeout", func(t *testing.T) {
		pinger := NewTCPPinger()
		defer pinger.Close()

		opts := &types.PingOptions{
			Count:   1,
			Timeout: 50 * time.Millisecond, // Short timeout
		}

		ctx := context.Background()
		// Use a port that is almost certainly not in use, this should cause a "connection refused" error
		result, err := pinger.Ping(ctx, "127.0.0.1:39876", opts)

		require.NoError(t, err) // The operation itself doesn't error, the result contains the status
		require.NotNil(t, result)
		assert.Equal(t, 1, result.Statistics.Sent)
		assert.Equal(t, 0, result.Statistics.Received)
		assert.Equal(t, 100.0, result.Statistics.LossRate)
		assert.Equal(t, types.StatusFailure, result.Status)
		assert.Equal(t, types.StatusFailure, result.Replies[0].Status)
		assert.Contains(t, result.Replies[0].Error, "connection refused")
	})

	// Case 3: Invalid hostname (DNS error)
	t.Run("InvalidHost", func(t *testing.T) {
		pinger := NewTCPPinger()
		defer pinger.Close()

		opts := &types.PingOptions{
			Count: 1,
		}

		ctx := context.Background()
		// A target with no host should be invalid
		_, err := pinger.Ping(ctx, ":1234", opts)
		require.Error(t, err)
		assert.ErrorIs(t, err, errors.ErrInvalidHost)
	})
}

func TestTCPPinger_PingStream(t *testing.T) {
	// Case 1: Successful stream to a live server
	t.Run("StreamSuccess", func(t *testing.T) {
		server, addr := setupTCPServer(t)
		defer server.Close()

		pinger := NewTCPPinger()
		defer pinger.Close()

		opts := &types.PingOptions{
			Count:    3,
			Interval: 50 * time.Millisecond,
			Timeout:  time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		replyChan, err := pinger.PingStream(ctx, addr, opts)
		require.NoError(t, err)

		receivedCount := 0
		for reply := range replyChan {
			require.NotNil(t, reply)
			if reply.Status == types.StatusSuccess {
				receivedCount++
			}
			fmt.Printf("Received reply: seq=%d, status=%v, rtt=%v\n", reply.Seq, reply.Status, reply.RTT)
		}

		assert.Equal(t, 3, receivedCount)
	})

	// Case 2: Stream with context cancellation
	t.Run("StreamCancel", func(t *testing.T) {
		server, addr := setupTCPServer(t)
		defer server.Close()

		pinger := NewTCPPinger()
		defer pinger.Close()

		opts := &types.PingOptions{
			Count:    10, // Attempt 10 pings
			Interval: 100 * time.Millisecond,
			Timeout:  time.Second,
		}

		// Create a context that cancels after 250ms (allowing ~2 pings)
		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		defer cancel()

		replyChan, err := pinger.PingStream(ctx, addr, opts)
		require.NoError(t, err)

		receivedCount := 0
		for range replyChan {
			receivedCount++
		}

		// Check that the stream was cancelled partway through
		assert.Greater(t, receivedCount, 0)
		assert.Less(t, receivedCount, 10)
	})
}
