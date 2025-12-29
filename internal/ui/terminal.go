package ui

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// ClearScreen 清空终端内容
func ClearScreen() {
	fmt.Print("\033[2J\033[H")
}

// TerminalSize 返回终端宽高（列、行）
func TerminalSize() (int, int, error) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 0, 0, err
	}
	return width, height, nil
}
