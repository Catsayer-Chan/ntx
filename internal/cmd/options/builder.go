package options

import (
	"github.com/catsayer/ntx/internal/app"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Builder 提供通用的命令选项构建流程：默认值 → 配置覆盖 → Flag 覆盖
type Builder[T any] struct {
	opts   T
	flags  *pflag.FlagSet
	appCtx *app.Context
}

// NewBuilder 创建新的选项构建器
func NewBuilder[T any](defaults T) *Builder[T] {
	return &Builder[T]{opts: defaults}
}

// WithCommand 使用命令的 Flag 集
func (b *Builder[T]) WithCommand(cmd *cobra.Command) *Builder[T] {
	if cmd != nil {
		b.flags = cmd.Flags()
	}
	return b
}

// WithFlags 直接注入 Flag 集
func (b *Builder[T]) WithFlags(flags *pflag.FlagSet) *Builder[T] {
	b.flags = flags
	return b
}

// WithContext 注入应用上下文
func (b *Builder[T]) WithContext(ctx *app.Context) *Builder[T] {
	b.appCtx = ctx
	return b
}

// ApplyConfig 使用提供的函数基于配置覆盖默认值
func (b *Builder[T]) ApplyConfig(apply func(opts T, ctx *app.Context)) *Builder[T] {
	if apply != nil && b.appCtx != nil {
		apply(b.opts, b.appCtx)
	}
	return b
}

// ApplyFlags 使用提供的函数基于命令行 Flag 覆盖选项
func (b *Builder[T]) ApplyFlags(apply func(opts T, flags *pflag.FlagSet)) *Builder[T] {
	if apply != nil && b.flags != nil {
		apply(b.opts, b.flags)
	}
	return b
}

// Result 返回最终构建的选项
func (b *Builder[T]) Result() T {
	return b.opts
}
