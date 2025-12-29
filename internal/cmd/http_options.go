package cmd

import (
	"time"

	"github.com/catsayer/ntx/internal/app"
	"github.com/catsayer/ntx/internal/cmd/options"
	"github.com/catsayer/ntx/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func buildHTTPOptions(cmd *cobra.Command, appCtx *app.Context) *types.HTTPOptions {
	defaults := &types.HTTPOptions{
		Timeout:        types.DefaultHTTPTimeout,
		FollowRedirect: true,
		MaxRedirects:   10,
	}
	return options.NewBuilder(defaults).
		WithContext(appCtx).
		WithCommand(cmd).
		ApplyConfig(func(opts *types.HTTPOptions, ctx *app.Context) {
			if ctx == nil || ctx.Config == nil {
				return
			}
			httpCfg := ctx.Config.HTTP
			if httpCfg.Timeout > 0 {
				opts.Timeout = httpCfg.Timeout
			}
			opts.FollowRedirect = httpCfg.FollowRedirect
			if httpCfg.MaxRedirects > 0 {
				opts.MaxRedirects = httpCfg.MaxRedirects
			}
			opts.UserAgent = httpCfg.UserAgent
		}).
		ApplyFlags(func(opts *types.HTTPOptions, flags *pflag.FlagSet) {
			if flags.Changed("timeout") {
				opts.Timeout = time.Duration(httpTimeout * float64(time.Second))
			}
			if flags.Changed("no-redirect") {
				opts.FollowRedirect = !httpNoRedirect
			}
		}).
		Result()
}
