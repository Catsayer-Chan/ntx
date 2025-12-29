package ping

import (
	"context"
	stderrors "errors"
	"fmt"

	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/pkg/types"
	"go.uber.org/zap"
)

// Mode 表示 ping 运行模式
type Mode int

const (
	// ModeStream 实时文本模式
	ModeStream Mode = iota
	// ModeMonitor 实时监控模式
	ModeMonitor
	// ModeBatch 批量结构化输出模式
	ModeBatch
)

// Config 控制运行参数
type Config struct {
	Mode         Mode
	OutputFormat types.OutputFormat
	NoColor      bool
}

// Runner 负责执行 ping 任务
type Runner struct {
	cfg     Config
	factory types.PingerFactory
}

// ErrPartialFailure 表示部分目标失败
var ErrPartialFailure = stderrors.New("partial ping failure")

// NewRunner 创建 Runner
func NewRunner(cfg Config, factory types.PingerFactory) *Runner {
	return &Runner{
		cfg:     cfg,
		factory: factory,
	}
}

// Run 根据配置执行 Ping
func (r *Runner) Run(ctx context.Context, targets []string, opts *types.PingOptions) error {
	if len(targets) == 0 {
		return fmt.Errorf("no targets specified")
	}
	if r.factory == nil {
		return fmt.Errorf("pinger factory is not configured")
	}

	targetOpts := *opts
	pinger, err := r.factory.Create(&targetOpts)
	if err != nil {
		logger.Error("创建 Pinger 失败", zap.Error(err))
		return err
	}
	defer pinger.Close()

	switch r.cfg.Mode {
	case ModeMonitor:
		return runPingMonitor(ctx, pinger, targets[0], &targetOpts)
	case ModeBatch:
		return runPingBatchConcurrent(ctx, pinger, targets, &targetOpts, r.cfg.OutputFormat, r.cfg.NoColor)
	default:
		return runPingStream(ctx, pinger, targets, &targetOpts, r.cfg.NoColor)
	}
}
