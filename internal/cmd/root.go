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
// - github.com/spf13/viper: 配置管理
// - internal/logger: 日志系统
// - internal/config: 配置管理
//
// 使用示例：
//   err := cmd.Execute()
//
// 作者: Catsayer
package cmd

import (
	"fmt"
	"os"

	"github.com/catsayer/ntx/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	cfgFile string
	verbose bool
	output  string
	noColor bool
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
	Version: "0.1.0",
}

// Execute 添加所有子命令到 root 命令并适当设置标志
// 这被 main.main() 调用，只需要对 rootCmd 发生一次
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// 全局持久化标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认: $HOME/.ntx.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "启用详细输出")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "text", "输出格式: text|json|yaml")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "禁用彩色输出")

	// 绑定到 viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color"))
}

// initConfig 读取配置文件和环境变量
func initConfig() {
	if cfgFile != "" {
		// 使用命令行指定的配置文件
		viper.SetConfigFile(cfgFile)
	} else {
		// 查找主目录
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// 在主目录中搜索 ".ntx" 配置文件
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".ntx")
	}

	// 读取环境变量
	viper.SetEnvPrefix("NTX")
	viper.AutomaticEnv()

	// 如果找到配置文件，则读取它
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Fprintln(os.Stderr, "使用配置文件:", viper.ConfigFileUsed())
		}
	}

	// 初始化日志系统
	logLevel := "info"
	if verbose {
		logLevel = "debug"
	}

	logConfig := logger.Config{
		Level:             logLevel,
		Development:       verbose,
		DisableCaller:     false,
		DisableStacktrace: !verbose,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
	}

	if err := logger.Init(logConfig); err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志系统失败: %v\n", err)
		os.Exit(1)
	}

	// 确保在程序退出时刷新日志
	defer logger.Sync()

	logger.Debug("配置初始化完成", zap.String("config", viper.ConfigFileUsed()))
}