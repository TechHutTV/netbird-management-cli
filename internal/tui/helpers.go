package tui

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// jsonDecode reads and decodes JSON from a reader into the target
func jsonDecode(r io.Reader, target interface{}) error {
	if err := json.NewDecoder(r).Decode(target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}

// formatBytes formats a byte count into a human-readable string (e.g. "1.5 KB")
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// formatLastSeen formats an ISO timestamp into a human-readable relative time
func formatLastSeen(ts string) string {
	if ts == "" || ts == "0001-01-01T00:00:00Z" {
		return "Never"
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		// Try without nanoseconds
		t, err = time.Parse("2006-01-02T15:04:05Z", ts)
		if err != nil {
			return ts
		}
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "Just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("2006-01-02")
	}
}
