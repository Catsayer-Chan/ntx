package ping

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/internal/output/formatter"
	"github.com/catsayer/ntx/pkg/types"
	"go.uber.org/zap"
)

func runPingBatchConcurrent(ctx context.Context, pinger types.Pinger, targets []string, opts *types.PingOptions, outputFormat types.OutputFormat, noColor bool) error {
	if len(targets) == 0 {
		return nil
	}

	concurrency := batchWorkerCount(len(targets))
	resultsChan := make(chan *types.PingResult, len(targets))
	jobs := make(chan string)

	var workers sync.WaitGroup
	workers.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer workers.Done()
			for t := range jobs {
				select {
				case <-ctx.Done():
					resultsChan <- &types.PingResult{
						Target: &types.Host{Hostname: t},
						Status: types.StatusFailure,
						Error:  ctx.Err(),
					}
					continue
				default:
				}

				targetOpts := *opts
				targetOpts.EnsurePort(t)

				result, err := pinger.Ping(ctx, t, &targetOpts)
				if err != nil {
					logger.Error("Ping 失败", zap.Error(err), zap.String("target", t))
					resultsChan <- &types.PingResult{
						Target: &types.Host{Hostname: t},
						Status: types.StatusFailure,
						Error:  err,
					}
					continue
				}
				resultsChan <- result
			}
		}()
	}

	for _, target := range targets {
		jobs <- target
	}
	close(jobs)

	workers.Wait()
	close(resultsChan)

	allResults := make([]*types.PingResult, 0, len(targets))
	allSuccess := true
	for res := range resultsChan {
		allResults = append(allResults, res)
		if res.Status == types.StatusFailure || (res.Statistics != nil && res.Statistics.Received == 0) {
			allSuccess = false
		}
	}

	f := formatter.NewFormatter(outputFormat, noColor)
	output, err := f.Format(allResults)
	if err != nil {
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	fmt.Print(output)

	if !allSuccess {
		return ErrPartialFailure
	}
	return nil
}

func batchWorkerCount(targetCount int) int {
	if targetCount <= 0 {
		return 1
	}

	concurrency := runtime.NumCPU()
	if concurrency < 8 {
		concurrency = 8
	}
	if concurrency > 32 {
		concurrency = 32
	}
	if concurrency > targetCount {
		concurrency = targetCount
	}
	if concurrency <= 0 {
		return 1
	}
	return concurrency
}
