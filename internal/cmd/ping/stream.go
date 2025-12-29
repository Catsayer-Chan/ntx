package ping

import (
	"context"
	stderrors "errors"
	"fmt"
	"os"
	"time"

	"github.com/catsayer/ntx/internal/logger"
	pkgerrors "github.com/catsayer/ntx/pkg/errors"
	"github.com/catsayer/ntx/pkg/stats"
	"github.com/catsayer/ntx/pkg/termutil"
	"github.com/catsayer/ntx/pkg/types"
	"go.uber.org/zap"
)

func runPingStream(ctx context.Context, pinger types.Pinger, targets []string, opts *types.PingOptions, noColor bool) error {
	printer := termutil.NewColorPrinter(noColor)

	for i, target := range targets {
		logger.Info("开始 Ping", zap.String("target", target), zap.String("protocol", string(opts.Protocol)))
		if err := streamSingleTarget(ctx, pinger, target, opts, printer); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		if i < len(targets)-1 {
			fmt.Println()
		}
	}
	return nil
}

func streamSingleTarget(ctx context.Context, pinger types.Pinger, target string, opts *types.PingOptions, printer *termutil.ColorPrinter) error {
	targetOpts := *opts
	targetOpts.EnsurePort(target)

	preflightOpts := targetOpts
	preflightOpts.Count = 1
	pingCtx, cancel := context.WithTimeout(ctx, targetOpts.Timeout)
	firstResult, err := pinger.Ping(pingCtx, target, &preflightOpts)
	cancel()
	if err != nil {
		var netErr *pkgerrors.NetworkError
		if stderrors.As(err, &netErr) && netErr.Op == "resolve" {
			return fmt.Errorf("ping: cannot resolve %s: Unknown host", target)
		}
		return err
	}
	targetHostname := target
	targetIP := target
	if firstResult != nil && firstResult.Target != nil {
		targetHostname = firstResult.Target.Hostname
		targetIP = firstResult.Target.IP
	}

	fmt.Printf("PING %s (%s) %d(%d) bytes of data.\n", targetHostname, targetIP, targetOpts.Size, targetOpts.Size+28)

	replyChan, err := pinger.PingStream(ctx, target, &targetOpts)
	if err != nil {
		return fmt.Errorf("错误: %w", err)
	}

	sent := 0
	received := 0
	var rtts []time.Duration
	var totalTime time.Duration
	startTime := time.Now()

	for reply := range replyChan {
		sent++
		if ctx.Err() != nil {
			break
		}
		if reply.Status == types.StatusSuccess {
			received++
			rtts = append(rtts, reply.RTT)
			fmt.Println(printer.Success(fmt.Sprintf("%d bytes from %s: icmp_seq=%d ttl=%d time=%.3f ms",
				reply.Bytes,
				targetIP,
				reply.Seq,
				reply.TTL,
				float64(reply.RTT.Microseconds())/1000.0,
			)))
		} else {
			fmt.Println(printer.Error(fmt.Sprintf("Request timeout for icmp_seq=%d", reply.Seq)))
		}
	}
	totalTime = time.Since(startTime)

	fmt.Printf("\n--- %s ping statistics ---\n", targetHostname)
	lossRate := 0.0
	if sent > 0 {
		lossRate = float64(sent-received) / float64(sent) * 100
	}
	fmt.Printf("%d packets transmitted, %d received, %.f%% packet loss, time %dms\n",
		sent, received, lossRate, totalTime.Milliseconds())

	if len(rtts) > 0 {
		min, max, avg, stddev := stats.ComputeRTTStats(rtts)
		fmt.Printf("rtt min/avg/max/mdev = %.3f/%.3f/%.3f/%.3f ms\n",
			float64(min.Microseconds())/1000.0,
			float64(avg.Microseconds())/1000.0,
			float64(max.Microseconds())/1000.0,
			float64(stddev.Microseconds())/1000.0)
	}

	return nil
}
