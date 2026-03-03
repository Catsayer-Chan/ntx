package batch

import (
	"context"
	"fmt"
	"time"

	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/pkg/types"
	"go.uber.org/zap"
)

func (e *Executor) executeScanTask(ctx context.Context, task Task, result *TaskResult) error {
	if len(task.Targets) == 0 {
		return fmt.Errorf("scan 任务未配置目标")
	}

	opts := types.DefaultScanOptions()

	if task.Options != nil {
		if timeout, ok := task.Options["timeout"].(int); ok {
			opts.Timeout = time.Duration(timeout) * time.Second
		}
		if concurrency, ok := task.Options["concurrency"].(int); ok {
			opts.Concurrency = concurrency
		}
		if ports, ok := task.Options["ports"].([]interface{}); ok {
			opts.Ports = make([]int, 0)
			for _, p := range ports {
				if port, ok := p.(int); ok {
					opts.Ports = append(opts.Ports, port)
				}
			}
		}
	}

	failures := 0
	for _, target := range task.Targets {
		scanResult, err := e.scanner.Scan(ctx, target, opts)
		if err != nil {
			logger.Error("扫描失败", zap.String("target", target), zap.Error(err))
			failures++
			continue
		}

		result.Results = append(result.Results, scanResult)

		logger.Info("扫描完成",
			zap.String("target", target),
			zap.Int("open_ports", scanResult.Summary.OpenPorts),
		)
	}

	if failures > 0 {
		return fmt.Errorf("scan 任务部分失败: %d/%d 个目标失败", failures, len(task.Targets))
	}
	if len(result.Results) == 0 {
		return fmt.Errorf("scan 任务未产生有效结果")
	}

	return nil
}
