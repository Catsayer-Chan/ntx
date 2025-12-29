package cmd

import (
	"encoding/json"
	"strings"
)

func buildHTTPHeaders(rawHeaders []string, body string) map[string]string {
	headers := make(map[string]string)
	for _, h := range rawHeaders {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key != "" {
			headers[key] = value
		}
	}

	if body == "" || headers["Content-Type"] != "" {
		return headers
	}

	var js json.RawMessage
	if json.Unmarshal([]byte(body), &js) == nil {
		headers["Content-Type"] = "application/json"
	} else {
		headers["Content-Type"] = "application/x-www-form-urlencoded"
	}

	return headers
}
