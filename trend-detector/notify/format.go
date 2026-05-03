package notify

import (
	"fmt"
	"time"

	"gnai/trend-detector/trend"
)

// FormatShiftAlert renders a shift alert message.
func FormatShiftAlert(st trend.State) string {
	prev := st.Previous.Trend
	if prev == "" {
		prev = trend.Unknown
	}
	return fmt.Sprintf(
		"TREND SHIFT — %s\nPrevious : %s (lasted %s, confidence %.0f%%)\nCurrent  : %s (confidence %.0f%%)\nMode     : %s\nReason   : %s\nTime     : %s",
		st.Current.Timestamp.Format("02-Jan-2006"),
		prev,
		formatDuration(time.Since(st.StartedAt)),
		st.Previous.Confidence*100,
		st.Current.Trend,
		st.Current.Confidence*100,
		st.Current.Mode,
		st.Current.Reasoning,
		st.Current.Timestamp.Format("15:04 MST"),
	)
}

// FormatDigest renders a periodic digest message.
func FormatDigest(st trend.State) string {
	return fmt.Sprintf(
		"TREND DIGEST — %s\nCurrent Trend : %s\nConfidence    : %.0f%%  |  Duration: %s\nMode          : %s\nReasoning     : %s",
		st.Current.Timestamp.Format("15:04 MST"),
		st.Current.Trend,
		st.Current.Confidence*100,
		formatDuration(time.Since(st.StartedAt)),
		st.Current.Mode,
		st.Current.Reasoning,
	)
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
