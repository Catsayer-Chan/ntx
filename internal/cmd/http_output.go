package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/catsayer/ntx/pkg/termutil"
	"github.com/catsayer/ntx/pkg/types"
)

func printHTTPResultText(result *types.HTTPResult, noColor bool) {
	printer := termutil.NewColorPrinter(noColor)
	bold := printer.Bold
	green := printer.Success
	yellow := printer.Warning
	red := printer.Error

	statusColor := green
	if result.StatusCode >= 400 {
		statusColor = red
	} else if result.StatusCode >= 300 {
		statusColor = yellow
	}

	fmt.Printf("%s %s %s\n",
		bold(result.Proto),
		statusColor(fmt.Sprintf("%d", result.StatusCode)),
		statusColor(result.Status))

	if httpIncludeHead || httpHeadOnly {
		for key, values := range result.Headers {
			for _, value := range values {
				fmt.Printf("%s: %s\n", bold(key), value)
			}
		}
		fmt.Println()
	}

	if !httpHeadOnly && len(result.Body) > 0 {
		if strings.Contains(result.ContentType, "application/json") {
			var js interface{}
			if json.Unmarshal(result.Body, &js) == nil {
				formatted, err := json.MarshalIndent(js, "", "  ")
				if err == nil {
					fmt.Println(string(formatted))
				} else {
					fmt.Println(string(result.Body))
				}
			} else {
				fmt.Println(string(result.Body))
			}
		} else {
			fmt.Println(string(result.Body))
		}
	}

	fmt.Printf("\n")
	fmt.Printf("Time: %v\n", result.Duration)
	fmt.Printf("Size: %s\n", formatSize(result.TransferredSize))
	if result.TLSUsed {
		fmt.Printf("TLS: %s\n", green("Yes"))
	}
}

func printHTTPBenchmarkText(result *types.HTTPBenchmarkResult, noColor bool) {
	printer := termutil.NewColorPrinter(noColor)
	bold := printer.Bold
	green := printer.Success
	red := printer.Error

	fmt.Printf("%s HTTP Benchmark Results\n", bold("==="))
	fmt.Printf("URL: %s\n", result.URL)
	fmt.Printf("Method: %s\n", result.Method)
	fmt.Println()

	fmt.Printf("Total Requests: %d\n", result.TotalRequests)
	fmt.Printf("Success: %s\n", green(fmt.Sprintf("%d (%.1f%%)",
		result.SuccessCount,
		float64(result.SuccessCount)/float64(result.TotalRequests)*100)))
	fmt.Printf("Failure: %s\n", red(fmt.Sprintf("%d (%.1f%%)",
		result.FailureCount,
		float64(result.FailureCount)/float64(result.TotalRequests)*100)))
	fmt.Println()

	fmt.Printf("Total Time: %v\n", result.TotalDuration)
	fmt.Printf("Min Time: %v\n", result.MinDuration)
	fmt.Printf("Max Time: %v\n", result.MaxDuration)
	fmt.Printf("Avg Time: %v\n", result.AvgDuration)
	fmt.Println()

	fmt.Printf("Requests/sec: %s\n", bold(fmt.Sprintf("%.2f", result.RequestsPerSec)))
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
