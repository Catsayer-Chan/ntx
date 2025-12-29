package formatter

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Table 提供简单的等宽表格渲染功能
type Table struct {
	headers   []string
	widths    []int
	rows      [][]string
	separator rune

	format     string
	lineLength int
}

// NewTable 创建一个新的表格渲染器
func NewTable(headers []string, widths []int) *Table {
	return (&Table{
		headers:   headers,
		widths:    widths,
		separator: '-',
	}).init()
}

// SetSeparator 设置表格分隔符
func (t *Table) SetSeparator(sep rune) {
	t.separator = sep
	t.lineLength = 0
}

// AddRow 添加一行数据
func (t *Table) AddRow(columns ...string) {
	row := make([]string, len(t.widths))
	for i := range row {
		if i < len(columns) {
			row[i] = columns[i]
		} else {
			row[i] = ""
		}
	}
	t.rows = append(t.rows, row)
}

// Render 输出表格
func (t *Table) Render(w io.Writer) {
	if w == nil {
		w = os.Stdout
	}

	if len(t.headers) > 0 {
		fmt.Fprintln(w, t.separatorLine())
		t.printRow(w, t.headers)
		fmt.Fprintln(w, t.separatorLine())
	}

	for _, row := range t.rows {
		t.printRow(w, row)
	}

	fmt.Fprintln(w, t.separatorLine())
}

func (t *Table) init() *Table {
	t.format = buildFormat(t.widths)
	t.lineLength = calcLineLength(t.widths)
	return t
}

func (t *Table) printRow(w io.Writer, row []string) {
	args := make([]any, len(t.widths))
	for i, col := range row {
		args[i] = col
	}
	fmt.Fprintf(w, t.format, args...)
}

func (t *Table) separatorLine() string {
	if t.lineLength == 0 {
		t.lineLength = calcLineLength(t.widths)
	}
	if t.lineLength <= 0 {
		return ""
	}
	return strings.Repeat(string(t.separator), t.lineLength)
}

func buildFormat(widths []int) string {
	if len(widths) == 0 {
		return "\n"
	}

	var b strings.Builder
	for i, width := range widths {
		if i > 0 {
			b.WriteByte(' ')
		}
		if width > 0 {
			b.WriteString(fmt.Sprintf("%%-%ds", width))
		} else {
			b.WriteString("%s")
		}
	}
	b.WriteByte('\n')
	return b.String()
}

func calcLineLength(widths []int) int {
	if len(widths) == 0 {
		return 0
	}

	total := len(widths) - 1 // spaces between columns
	for _, width := range widths {
		if width > 0 {
			total += width
		} else {
			total += 5
		}
	}
	return total
}
