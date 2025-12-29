package batch

import (
	"context"
	"time"

	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/pkg/types"
	"go.uber.org/zap"
)

func (e *Executor) executeScanTask(ctx context.Context, task Task, result *TaskResult) error {
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

	for _, target := range task.Targets {
		scanResult, err := e.scanner.Scan(ctx, target, opts)
		if err != nil {
			logger.Error("扫描失败", zap.String("target", target), zap.Error(err))
			continue
		}

		result.Results = append(result.Results, scanResult)

		logger.Info("扫描完成",
			zap.String("target", target),
			zap.Int("open_ports", scanResult.Summary.OpenPorts),
		)
	}

	return nil
}
