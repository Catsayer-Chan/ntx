package batch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/pkg/types"
	"go.uber.org/zap"
)

func (e *Executor) executePingTask(ctx context.Context, task Task, result *TaskResult) error {
	if len(task.Targets) == 0 {
		return fmt.Errorf("ping 任务未配置目标")
	}

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
	failures := 0
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
				mu.Lock()
				failures++
				mu.Unlock()
				return
			}
			if pingResult == nil {
				logger.Error("Ping 返回空结果", zap.String("target", t))
				mu.Lock()
				failures++
				mu.Unlock()
				return
			}

			mu.Lock()
			result.Results = append(result.Results, pingResult)
			if pingResult.Status != types.StatusSuccess {
				failures++
			}
			mu.Unlock()

			if pingResult.Statistics != nil {
				logger.Info("Ping 完成",
					zap.String("target", t),
					zap.String("status", string(pingResult.Status)),
					zap.Duration("avg_rtt", pingResult.Statistics.AvgRTT),
					zap.Float64("loss", pingResult.Statistics.LossRate),
				)
			} else {
				logger.Info("Ping 完成",
					zap.String("target", t),
					zap.String("status", string(pingResult.Status)),
				)
			}
		}(target)
	}

	wg.Wait()

	if failures > 0 {
		return fmt.Errorf("ping 任务部分失败: %d/%d 个目标失败", failures, len(task.Targets))
	}
	if len(result.Results) == 0 {
		return fmt.Errorf("ping 任务未产生有效结果")
	}

	return nil
}
