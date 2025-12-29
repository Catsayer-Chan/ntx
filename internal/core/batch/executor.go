package batch

import (
	"context"
	"time"

	"github.com/catsayer/ntx/internal/core/dns"
	"github.com/catsayer/ntx/internal/core/ping"
	"github.com/catsayer/ntx/internal/core/scan"
	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/pkg/types"
	"go.uber.org/zap"
)

// Executor 任务执行器
type Executor struct {
	pinger   types.Pinger
	resolver *dns.Resolver
	scanner  *scan.TCPScanner
}

// NewExecutor 创建新的任务执行器
func NewExecutor() *Executor {
	opts := types.DefaultPingOptions()
	var pinger types.Pinger
	icmpPinger, err := ping.NewICMPPinger(opts)
	if err != nil {
		pinger = ping.NewTCPPinger(opts)
	} else {
		pinger = icmpPinger
	}

	dnsOpts := &types.DNSOptions{
		Server:  types.DefaultDNSServer,
		Timeout: types.DefaultDNSTimeout,
	}

	return &Executor{
		pinger:   pinger,
		resolver: dns.NewResolver(dnsOpts),
		scanner:  scan.NewTCPScanner(),
	}
}

// ExecuteFile 执行配置文件中的任务
func (e *Executor) ExecuteFile(ctx context.Context, configFile string) (*BatchResult, error) {
	logger.Info("加载任务配置文件", zap.String("file", configFile))

	tasks, err := loadTaskConfig(configFile)
	if err != nil {
		return nil, err
	}

	return e.ExecuteTasks(ctx, tasks)
}

// ExecuteTasks 执行任务列表
func (e *Executor) ExecuteTasks(ctx context.Context, tasks []Task) (*BatchResult, error) {
	startTime := time.Now()

	logger.Info("开始执行批量任务", zap.Int("count", len(tasks)))

	result := &BatchResult{
		TotalTasks:  len(tasks),
		TaskResults: make([]*TaskResult, 0),
	}

	for _, task := range tasks {
		if !task.Enabled {
			logger.Info("跳过禁用的任务", zap.String("name", task.Name))
			continue
		}

		taskResult := e.executeTask(ctx, task)
		result.TaskResults = append(result.TaskResults, taskResult)

		if taskResult.Success {
			result.SuccessTasks++
		} else {
			result.FailedTasks++
		}
	}

	result.TotalDuration = time.Since(startTime)

	logger.Info("批量任务执行完成",
		zap.Int("total", result.TotalTasks),
		zap.Int("success", result.SuccessTasks),
		zap.Int("failed", result.FailedTasks),
		zap.Duration("duration", result.TotalDuration),
	)

	return result, nil
}

func (e *Executor) executeTask(ctx context.Context, task Task) *TaskResult {
	startTime := time.Now()

	logger.Info("执行任务", zap.String("name", task.Name), zap.String("type", string(task.Type)))

	result := &TaskResult{
		TaskName:  task.Name,
		TaskType:  task.Type,
		StartTime: startTime,
		Results:   make([]interface{}, 0),
	}

	var err error

	switch task.Type {
	case TaskTypePing:
		err = e.executePingTask(ctx, task, result)
	case TaskTypeDNS:
		err = e.executeDNSTask(ctx, task, result)
	case TaskTypeScan:
		err = e.executeScanTask(ctx, task, result)
	default:
		err = ErrUnsupportedTaskType(task.Type)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Success = err == nil
	result.Error = err

	return result
}
