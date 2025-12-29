package cmd

import (
	"fmt"
	"os"

	"github.com/catsayer/ntx/internal/app"
	"github.com/catsayer/ntx/pkg/types"
	"github.com/spf13/cobra"
)

func mustAppContext(cmd *cobra.Command) *app.Context {
	if cmd == nil {
		fmt.Fprintln(os.Stderr, "命令未初始化")
		os.Exit(1)
	}

	if ctx, ok := app.FromContext(cmd.Context()); ok && ctx != nil {
		return ctx
	}

	if appCtx != nil {
		newCtx := app.WithContext(cmd.Context(), appCtx)
		cmd.SetContext(newCtx)
		return appCtx
	}

	fmt.Fprintln(os.Stderr, "应用上下文未初始化")
	os.Exit(1)
	return nil
}

func outputFormatFromCmd(cmd *cobra.Command) types.OutputFormat {
	return types.OutputFormat(mustAppContext(cmd).Flags.Output)
}

func noColorFromCmd(cmd *cobra.Command) bool {
	return mustAppContext(cmd).Flags.NoColor
}

func verboseEnabled(cmd *cobra.Command) bool {
	return mustAppContext(cmd).Flags.Verbose
}
