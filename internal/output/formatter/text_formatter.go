package formatter

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// TextFormatter 提供简单的 CLI 文本输出能力
type TextFormatter struct {
	colorEnabled bool
	width        int
}

const defaultDividerWidth = 70

// NewTextFormatter 创建文本格式化器
func NewTextFormatter(colorEnabled bool) *TextFormatter {
	return &TextFormatter{
		colorEnabled: colorEnabled,
		width:        defaultDividerWidth,
	}
}

// PrintHeader 打印标题和分隔线
func (f *TextFormatter) PrintHeader(title string) {
	fmt.Println(f.divider("="))
	fmt.Println(f.styledTitle(title, true))
	fmt.Println(f.divider("="))
}

// PrintSubHeader 打印子标题
func (f *TextFormatter) PrintSubHeader(title string) {
	fmt.Println(f.styledTitle(title, false))
}

func (f *TextFormatter) divider(char string) string {
	if f.width <= 0 {
		f.width = defaultDividerWidth
	}
	return strings.Repeat(char, f.width)
}

func (f *TextFormatter) styledTitle(title string, bold bool) string {
	if !f.colorEnabled {
		return fmt.Sprintf("%s", title)
	}

	attrs := []color.Attribute{color.FgCyan}
	if bold {
		attrs = append(attrs, color.Bold)
	}
	return color.New(attrs...).Sprint(title)
}
