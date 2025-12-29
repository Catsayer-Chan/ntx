package types

// Common table widths/dividers used across formatters to keep layout consistent.
const (
	// Ping output
	TableWidthPingText  = 60
	TableWidthPingTable = 78

	// Trace output
	TableWidthTraceText  = 70
	TableWidthTraceTable = 95

	// Misc UI sections
	DividerWidthConnStats  = 40
	DividerWidthBatchTasks = 60
)

// Column widths for formatted table output.
const (
	ColumnWidthPingSeq    = 6
	ColumnWidthPingFrom   = 20
	ColumnWidthPingBytes  = 10
	ColumnWidthPingTTL    = 10
	ColumnWidthPingTime   = 10
	ColumnWidthPingStatus = 10

	ColumnWidthTraceHop   = 4
	ColumnWidthTraceHost  = 40
	ColumnWidthTraceProbe = 12
	ColumnWidthTraceAvg   = 8
)
