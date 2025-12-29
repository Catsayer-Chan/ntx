package batch

import (
	"context"
	"sync"
	"time"

	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/pkg/types"
	"go.uber.org/zap"
)

func (e *Executor) executePingTask(ctx context.Context, task Task, result *TaskResult) error {
	opts := types.DefaultPingOptions()

	if task.Options != nil {
		if count, ok := task.Options["count"].(int); ok {
			opts.Count = count
		}
		if timeout, ok := task.Options["timeout"].(int); ok {
			opts.Timeout = time.Duration(timeout) * time.Second
		}
	}

	concurrency := task.Concurrency
	if concurrency == 0 {
		concurrency = 10
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, concurrency)

	for _, target := range task.Targets {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			pingResult, err := e.pinger.Ping(ctx, t, opts)
			if err != nil {
				logger.Error("Ping 失败", zap.String("target", t), zap.Error(err))
				return
			}

			mu.Lock()
			result.Results = append(result.Results, pingResult)
			mu.Unlock()

			logger.Info("Ping 完成",
				zap.String("target", t),
				zap.Duration("avg_rtt", pingResult.Statistics.AvgRTT),
				zap.Float64("loss", pingResult.Statistics.LossRate),
			)
		}(target)
	}

	wg.Wait()
	return nil
}
