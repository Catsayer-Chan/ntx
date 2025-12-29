package ping

import (
	"context"
	"fmt"
	"time"

	"github.com/catsayer/ntx/internal/ui"
	"github.com/catsayer/ntx/pkg/types"
	"github.com/guptarohit/asciigraph"
)

func runPingMonitor(ctx context.Context, pinger types.Pinger, target string, opts *types.PingOptions) error {
	targetOpts := *opts
	targetOpts.EnsurePort(target)

	pingCtx, cancel := context.WithTimeout(ctx, targetOpts.Timeout)
	firstResult, err := pinger.Ping(pingCtx, target, &types.PingOptions{Count: 1, Timeout: targetOpts.Timeout})
	cancel()
	if err != nil || (firstResult != nil && firstResult.Statistics.Received == 0) {
		return fmt.Errorf("ping: cannot resolve %s: Unknown host", target)
	}
	targetHostname := firstResult.Target.Hostname
	targetIP := firstResult.Target.IP

	replyChan, err := pinger.PingStream(ctx, target, &targetOpts)
	if err != nil {
		return err
	}

	var rtts []float64
	sent := 0
	received := 0
	var lastRTT, minRTT, maxRTT, avgRTT time.Duration

	ui.ClearScreen()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n监控结束。")
			return nil
		case reply, ok := <-replyChan:
			if !ok {
				fmt.Println("\n监控完成。")
				return nil
			}

			sent++
			if reply.Status == types.StatusSuccess {
				received++
				rttMs := float64(reply.RTT.Microseconds()) / 1000.0
				rtts = append(rtts, rttMs)
				lastRTT = reply.RTT

				if minRTT == 0 || reply.RTT < minRTT {
					minRTT = reply.RTT
				}
				if reply.RTT > maxRTT {
					maxRTT = reply.RTT
				}
				avgRTT = (avgRTT*time.Duration(received-1) + reply.RTT) / time.Duration(received)
			} else {
				rtts = append(rtts, 0)
				lastRTT = 0
			}

			width, _, err := ui.TerminalSize()
			if err != nil || width <= 0 {
				width = 80
			}
			maxDataPoints := width - 15
			if maxDataPoints < 10 {
				maxDataPoints = 10
			}
			if len(rtts) > maxDataPoints {
				rtts = rtts[len(rtts)-maxDataPoints:]
			}

			graph := asciigraph.Plot(rtts, asciigraph.Height(10), asciigraph.Width(width-15), asciigraph.Caption("RTT (ms)"))

			lossRate := 0.0
			if sent > 0 {
				lossRate = float64(sent-received) / float64(sent) * 100
			}
			statsLine := fmt.Sprintf(
				"Target: %s (%s) | Sent: %d | Received: %d | Loss: %.1f%% | Last: %v | Avg: %v | Min: %v | Max: %v",
				targetHostname, targetIP, sent, received, lossRate,
				lastRTT.Round(time.Microsecond),
				avgRTT.Round(time.Microsecond),
				minRTT.Round(time.Microsecond),
				maxRTT.Round(time.Microsecond),
			)

			ui.ClearScreen()
			fmt.Println(statsLine)
			fmt.Println(graph)
		}
	}
}
