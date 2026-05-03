package backtest

import (
	"fmt"
	"strings"
	"time"

	"gnai/trend-detector/trend"
)

// Period represents one continuous trend in the backtest.
type Period struct {
	Trend      trend.TrendType
	Start      time.Time
	End        time.Time
	Duration   time.Duration
	CandleSpan int
}

// Report summarises a completed backtest run.
type Report struct {
	Symbol        string
	From          time.Time
	To            time.Time
	Interval      string
	TotalCandles  int
	ShiftCount    int
	AvgConfidence float64
	Periods       []Period
}

// Print outputs the report as a human-readable text block.
func (r Report) Print() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("=== BACKTEST REPORT ===\n"))
	sb.WriteString(fmt.Sprintf("Symbol    : %s\n", r.Symbol))
	sb.WriteString(fmt.Sprintf("Period    : %s → %s\n", r.From.Format("2006-01-02"), r.To.Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("Interval  : %s\n", r.Interval))
	sb.WriteString(fmt.Sprintf("Candles   : %d\n", r.TotalCandles))
	sb.WriteString(fmt.Sprintf("Shifts    : %d\n", r.ShiftCount))
	sb.WriteString(fmt.Sprintf("Avg Conf  : %.1f%%\n\n", r.AvgConfidence*100))

	// Trend distribution
	dist := map[trend.TrendType]int{}
	totalDur := map[trend.TrendType]time.Duration{}
	for _, p := range r.Periods {
		dist[p.Trend]++
		totalDur[p.Trend] += p.Duration
	}
	sb.WriteString("--- Trend Distribution ---\n")
	for _, t := range []trend.TrendType{trend.Bullish, trend.Bearish, trend.Sideways, trend.HighVolatility} {
		if dist[t] > 0 {
			sb.WriteString(fmt.Sprintf("  %-16s %3d periods  avg duration %s\n",
				t, dist[t], avgDur(totalDur[t], dist[t])))
		}
	}

	sb.WriteString("\n--- Trend Timeline ---\n")
	for _, p := range r.Periods {
		sb.WriteString(fmt.Sprintf("  %s  %-16s  %s → %s  (%s)\n",
			p.Start.Format("Jan 02 15:04"),
			p.Trend,
			p.Start.Format("15:04"),
			p.End.Format("15:04"),
			fmtDuration(p.Duration),
		))
	}

	return sb.String()
}

func avgDur(total time.Duration, count int) string {
	if count == 0 {
		return "—"
	}
	return fmtDuration(total / time.Duration(count))
}

func fmtDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%02dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
