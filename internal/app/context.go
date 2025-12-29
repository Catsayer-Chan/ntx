package app

import (
	"context"

	"github.com/catsayer/ntx/internal/config"
	"github.com/catsayer/ntx/internal/core/ping"
	"github.com/catsayer/ntx/pkg/types"
)

type contextKey struct{}

// GlobalFlags 保存根命令的全局标志
type GlobalFlags struct {
	Verbose bool
	Output  string
	NoColor bool
}

// Context 聚合配置和依赖
type Context struct {
	Config      *config.Config
	Flags       GlobalFlags
	PingFactory types.PingerFactory
}

// NewContext 构建默认应用上下文
func NewContext(cfg *config.Config, flags GlobalFlags) *Context {
	ctx := &Context{
		Config: cfg,
		Flags:  flags,
	}
	ctx.PingFactory = ping.NewFactory()
	return ctx
}

// WithContext 将应用上下文注入标准 context
func WithContext(parent context.Context, appCtx *Context) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithValue(parent, contextKey{}, appCtx)
}

// FromContext 从标准 context 读取应用上下文
func FromContext(ctx context.Context) (*Context, bool) {
	if ctx == nil {
		return nil, false
	}
	appCtx, ok := ctx.Value(contextKey{}).(*Context)
	if !ok || appCtx == nil {
		return nil, false
	}
	return appCtx, true
}
