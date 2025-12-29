package termutil

import (
	"fmt"

	"github.com/fatih/color"
)

// ColorPrinter 提供常用的彩色输出函数
type ColorPrinter struct {
	Success func(...interface{}) string
	Error   func(...interface{}) string
	Warning func(...interface{}) string
	Info    func(...interface{}) string
	Bold    func(...interface{}) string
	Muted   func(...interface{}) string
}

// NewColorPrinter 根据 noColor 标志创建统一的彩色输出器
func NewColorPrinter(noColor bool) *ColorPrinter {
	color.NoColor = noColor

	return &ColorPrinter{
		Success: sprint(noColor, color.FgGreen),
		Error:   sprint(noColor, color.FgRed),
		Warning: sprint(noColor, color.FgYellow),
		Info:    sprint(noColor, color.FgCyan),
		Bold:    sprint(noColor, color.Bold),
		Muted:   sprint(noColor, color.FgHiBlack),
	}
}

func sprint(noColor bool, attrs ...color.Attribute) func(...interface{}) string {
	if noColor {
		return fmt.Sprint
	}
	return color.New(attrs...).SprintFunc()
}
