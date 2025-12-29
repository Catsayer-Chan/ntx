package buildinfo

import "fmt"

var (
	// Version 编译时注入的版本号
	Version = "dev"
	// GitCommit 编译时注入的 git commit
	GitCommit = "unknown"
	// BuildTime 编译时间
	BuildTime = "unknown"
)

// UserAgent 返回统一的 User-Agent 字符串
func UserAgent() string {
	return fmt.Sprintf("NTX/%s", Version)
}

// FullVersion 返回包含额外信息的版本描述
func FullVersion() string {
	switch {
	case GitCommit != "unknown" && BuildTime != "unknown":
		return fmt.Sprintf("%s (%s, %s)", Version, GitCommit, BuildTime)
	case GitCommit != "unknown":
		return fmt.Sprintf("%s (%s)", Version, GitCommit)
	case BuildTime != "unknown":
		return fmt.Sprintf("%s (%s)", Version, BuildTime)
	default:
		return Version
	}
}
