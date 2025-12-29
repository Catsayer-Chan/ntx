package cmd

import (
	"time"

	"github.com/catsayer/ntx/internal/app"
	"github.com/catsayer/ntx/internal/cmd/options"
	"github.com/catsayer/ntx/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func buildDNSOptions(cmd *cobra.Command, appCtx *app.Context) *types.DNSOptions {
	defaults := &types.DNSOptions{
		Server:  types.DefaultDNSServer,
		Timeout: types.DefaultDNSTimeout,
	}
	return options.NewBuilder(defaults).
		WithContext(appCtx).
		WithCommand(cmd).
		ApplyConfig(func(opts *types.DNSOptions, ctx *app.Context) {
			if ctx == nil || ctx.Config == nil {
				return
			}
			if ctx.Config.DNS.Server != "" {
				opts.Server = ctx.Config.DNS.Server
			}
			if ctx.Config.DNS.Timeout > 0 {
				opts.Timeout = ctx.Config.DNS.Timeout
			}
		}).
		ApplyFlags(func(opts *types.DNSOptions, flags *pflag.FlagSet) {
			if flags.Changed("server") {
				opts.Server = dnsServer
			}
			if flags.Changed("timeout") {
				opts.Timeout = time.Duration(dnsTimeout * float64(time.Second))
			}
		}).
		Result()
}
