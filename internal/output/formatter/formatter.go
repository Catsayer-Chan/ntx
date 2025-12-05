// Package formatter 提供输出格式化功能
//
// 本模块支持多种输出格式：
// - Text: 人类可读的文本格式
// - JSON: JSON 格式
// - YAML: YAML 格式
// - Table: 表格格式
//
// 依赖：
// - encoding/json: JSON 编码
// - gopkg.in/yaml.v3: YAML 编码
//
// 使用示例：
//
//	formatter := formatter.NewFormatter(types.OutputJSON, false)
//	output, err := formatter.Format(result)
//
// 作者: Catsayer
package formatter

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/catsayer/ntx/pkg/types"
	"gopkg.in/yaml.v3"
)

// Formatter 格式化器接口
type Formatter interface {
	// Format 格式化数据
	Format(data interface{}) (string, error)
	// FormatTo 格式化数据并写入 Writer
	FormatTo(w io.Writer, data interface{}) error
}

// Config 格式化器配置
type Config struct {
	// Format 输出格式
	Format types.OutputFormat
	// NoColor 是否禁用颜色
	NoColor bool
	// Indent 是否缩进（JSON/YAML）
	Indent bool
}

// formatter 格式化器实现
type formatter struct {
	config Config
}

// NewFormatter 创建格式化器
func NewFormatter(format types.OutputFormat, noColor bool) Formatter {
	return &formatter{
		config: Config{
			Format:  format,
			NoColor: noColor,
			Indent:  true,
		},
	}
}

// NewFormatterWithConfig 使用配置创建格式化器
func NewFormatterWithConfig(config Config) Formatter {
	return &formatter{
		config: config,
	}
}

// Format 格式化数据
func (f *formatter) Format(data interface{}) (string, error) {
	switch f.config.Format {
	case types.OutputJSON:
		return f.formatJSON(data)
	case types.OutputYAML:
		return f.formatYAML(data)
	case types.OutputText:
		return f.formatText(data)
	case types.OutputTable:
		return f.formatTable(data)
	default:
		return "", fmt.Errorf("unsupported output format: %s", f.config.Format)
	}
}

// FormatTo 格式化数据并写入 Writer
func (f *formatter) FormatTo(w io.Writer, data interface{}) error {
	output, err := f.Format(data)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(output))
	return err
}

// formatJSON 格式化为 JSON
func (f *formatter) formatJSON(data interface{}) (string, error) {
	var b []byte
	var err error

	if f.config.Indent {
		b, err = json.MarshalIndent(data, "", "  ")
	} else {
		b, err = json.Marshal(data)
	}

	if err != nil {
		return "", fmt.Errorf("json marshal failed: %w", err)
	}

	return string(b), nil
}

// formatYAML 格式化为 YAML
func (f *formatter) formatYAML(data interface{}) (string, error) {
	b, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("yaml marshal failed: %w", err)
	}
	return string(b), nil
}

// formatText 格式化为文本
func (f *formatter) formatText(data interface{}) (string, error) {
	// 根据数据类型选择合适的文本格式化器
	switch v := data.(type) {
	case *types.PingResult:
		return FormatPingText(v, f.config.NoColor), nil
	case *types.TraceResult:
		return FormatTraceText(v, f.config.NoColor), nil
	default:
		// 默认使用 JSON 格式
		return f.formatJSON(data)
	}
}

// formatTable 格式化为表格
func (f *formatter) formatTable(data interface{}) (string, error) {
	// 根据数据类型选择合适的表格格式化器
	switch v := data.(type) {
	case *types.PingResult:
		return FormatPingTable(v, f.config.NoColor), nil
	case *types.TraceResult:
		return FormatTraceTable(v, f.config.NoColor), nil
	default:
		// 默认使用文本格式
		return f.formatText(data)
	}
}
