// Package cmd 提供 NTX 的命令行接口实现
//
// 本模块基于 Cobra 框架实现命令行接口，提供：
// - 根命令和子命令管理
// - 全局参数配置
// - 命令执行流程控制
// - 帮助信息生成
//
// 依赖：
// - github.com/spf13/cobra: CLI 框架
// - internal/logger: 日志系统
//
// 使用示例：
//
//	err := cmd.Execute()
//
// 作者: Catsayer
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/catsayer/ntx/internal/app"
	"github.com/catsayer/ntx/internal/config"
	"github.com/catsayer/ntx/internal/logger"
	"github.com/catsayer/ntx/pkg/buildinfo"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	globalFlags = app.GlobalFlags{
		Output: "text",
	}
	configFile string
	appCtx     *app.Context
)

// rootCmd 表示在没有任何子命令时调用的基本命令
var rootCmd = &cobra.Command{
	Use:   "ntx",
	Short: "NTX - 现代化的网络调试工具集",
	Long: `NTX (Network Tools eXtended) 是一个现代化的网络调试命令整合工具。

整合了 ping、traceroute、netstat、ss、ifconfig 等常用网络调试命令，
提供统一的命令行接口和一致的输出格式。

特性：
  • 多协议支持（ICMP/TCP/HTTP）
  • 跨平台一致性（Linux/macOS/Windows）
  • 智能诊断功能
  • 批量任务处理
  • 插件化架构
  • 结构化输出（JSON/YAML/Table）

示例：
  ntx ping google.com -c 5
  ntx trace google.com --max-hops 20
  ntx scan 192.168.1.1 -p 1-1024
  ntx diag`,
	Version: buildinfo.Version,
}

// Execute 添加所有子命令到 root 命令并适当设置标志
// 这被 main.main() 调用，只需要对 rootCmd 发生一次
func Execute() error {
	defer logger.Sync()
	rootContext := rootCmd.Context()
	if rootContext == nil {
		rootContext = context.Background()
	}
	return rootCmd.ExecuteContext(rootContext)
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentPreRun = injectAppContext

	// 全局持久化标志
	rootCmd.PersistentFlags().BoolVarP(&globalFlags.Verbose, "verbose", "v", false, "启用详细输出")
	rootCmd.PersistentFlags().StringVarP(&globalFlags.Output, "output", "o", "text", "输出格式: text|json|yaml")
	rootCmd.PersistentFlags().BoolVar(&globalFlags.NoColor, "no-color", false, "禁用彩色输出")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "配置文件路径 (默认自动搜索)")
}

// initConfig 初始化配置和日志系统
func initConfig() {
	loader := config.NewLoader()
	cfg, usedPath, err := loader.LoadWithEnv(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}
	if err := config.Validate(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "配置验证失败: %v\n", err)
		os.Exit(1)
	}

	flags := rootCmd.PersistentFlags()
	if !flags.Changed("verbose") {
		globalFlags.Verbose = cfg.Global.Verbose
	}
	if !flags.Changed("output") && cfg.Global.Output != "" {
		globalFlags.Output = cfg.Global.Output
	}
	if !flags.Changed("no-color") {
		globalFlags.NoColor = cfg.Global.NoColor
	}
	if globalFlags.Output == "" {
		globalFlags.Output = "text"
	}

	logConfig := logger.Config{
		Level:             cfg.Global.LogLevel,
		Development:       globalFlags.Verbose,
		DisableCaller:     false,
		DisableStacktrace: !globalFlags.Verbose,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
	}
	if cfg.Global.LogFile != "" {
		logConfig.OutputPaths = append(logConfig.OutputPaths, cfg.Global.LogFile)
		logConfig.ErrorOutputPaths = append(logConfig.ErrorOutputPaths, cfg.Global.LogFile)
	}

	if err := logger.Init(logConfig); err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志系统失败: %v\n", err)
		os.Exit(1)
	}

	appCtx = app.NewContext(cfg, globalFlags)
	rootContext := app.WithContext(rootCmd.Context(), appCtx)
	rootCmd.SetContext(rootContext)

	if usedPath != "" {
		logger.Debug("配置初始化完成", zap.String("config", usedPath))
	} else {
		logger.Debug("配置初始化完成，使用内置默认配置")
	}
}

func injectAppContext(cmd *cobra.Command, _ []string) {
	if appCtx == nil {
		fmt.Fprintln(os.Stderr, "应用上下文未初始化")
		os.Exit(1)
	}
	if ctx, ok := app.FromContext(cmd.Context()); ok && ctx != nil {
		return
	}
	cmd.SetContext(app.WithContext(cmd.Context(), appCtx))
}
