package metrics

import (
	"fmt"
	"strings"
	"time"
)

func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func FormatUptime(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func (m *Metrics) Summary() string {
	sent, recv := m.GetBandwidth()
	httpTotal, avgLatency, minLatency, maxLatency := m.GetHTTPStats()
	uptime := m.GetUptime()

	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("Metrics Summary\n")
	sb.WriteString("─────────────────────────────────────────────────────────────\n")

	sb.WriteString(fmt.Sprintf("Active Streams     %d\n", m.GetActiveStreams()))
	sb.WriteString(fmt.Sprintf("Total Streams      %d\n", m.GetTotalStreams()))
	sb.WriteString(fmt.Sprintf("Total Connections  %d\n\n", m.TotalConnections))

	sb.WriteString(fmt.Sprintf("Data Sent          %s\n", FormatBytes(sent)))
	sb.WriteString(fmt.Sprintf("Data Received      %s\n", FormatBytes(recv)))
	sb.WriteString(fmt.Sprintf("Total Transfer     %s\n\n", FormatBytes(sent+recv)))

	if httpTotal > 0 {
		sb.WriteString(fmt.Sprintf("HTTP Requests      %d\n", httpTotal))
		sb.WriteString(fmt.Sprintf("Avg Latency        %dms\n", avgLatency.Milliseconds()))
		sb.WriteString(fmt.Sprintf("Min Latency        %dms\n", minLatency.Milliseconds()))
		sb.WriteString(fmt.Sprintf("Max Latency        %dms\n\n", maxLatency.Milliseconds()))

		codes := m.GetStatusCodeCounts()
		if len(codes) > 0 {
			sb.WriteString("Status Codes\n")
			for code, count := range codes {
				sb.WriteString(fmt.Sprintf("  %d: %d requests\n", code, count))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString(fmt.Sprintf("Uptime             %s\n", FormatUptime(uptime)))

	return sb.String()
}

func (m *Metrics) OneLiner() string {
	active := m.GetActiveStreams()
	total := m.GetTotalStreams()
	sent, recv := m.GetBandwidth()
	httpTotal, avgLatency, _, _ := m.GetHTTPStats()

	if httpTotal > 0 {
		return fmt.Sprintf("Streams: %d/%d | Data: ↑%s ↓%s | HTTP: %d req, %dms avg",
			active, total, FormatBytes(sent), FormatBytes(recv), httpTotal, avgLatency.Milliseconds())
	}

	return fmt.Sprintf("Streams: %d/%d | Data: ↑%s ↓%s",
		active, total, FormatBytes(sent), FormatBytes(recv))
}
