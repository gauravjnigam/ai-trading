package trend

import "time"

// TrendType represents the classified market trend.
type TrendType string

const (
	Bullish       TrendType = "BULLISH"
	Bearish       TrendType = "BEARISH"
	Sideways      TrendType = "SIDEWAYS"
	HighVolatility TrendType = "HIGH_VOLATILITY"
	Unknown        TrendType = "UNKNOWN"
)

// Classification is the output of a single trend detection run.
type Classification struct {
	Trend      TrendType
	Confidence float64 // 0.0–1.0
	Reasoning  string
	Timestamp  time.Time
	Mode       string // "rule-based" or "claude-enhanced"
}

// State tracks the current confirmed trend and its history.
type State struct {
	Current   Classification
	Previous  Classification
	StartedAt time.Time
	Shifted   bool // true on the tick when a shift was just confirmed
}

// Duration returns how long the current trend has been active.
func (s State) Duration() time.Duration {
	return time.Since(s.StartedAt)
}
