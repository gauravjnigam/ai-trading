package backtest

import (
	"context"
	"fmt"
	"log"
	"time"

	"gnai/trend-detector/indicator"
	"gnai/trend-detector/market"
	"gnai/trend-detector/trend"
)

// Config holds backtesting parameters.
type Config struct {
	Symbol           string
	Exchange         string
	From             time.Time
	To               time.Time
	KiteInterval     string // e.g. "5minute"
	StoreLookback    int
	ConfidenceThresh float64
}

// Run replays historical candles through the indicator + classifier pipeline
// and returns a Report.
func Run(ctx context.Context, candles []market.OHLCV, classifier *trend.Classifier, cfg Config) Report {
	store := market.NewStore(cfg.StoreLookback, 5*time.Minute)

	report := Report{
		Symbol:   cfg.Symbol,
		From:     cfg.From,
		To:       cfg.To,
		Interval: cfg.KiteInterval,
	}

	var currentTrend trend.TrendType = trend.Unknown
	var trendStart time.Time
	var allConf []float64

	for i, candle := range candles {
		select {
		case <-ctx.Done():
			return report
		default:
		}

		store.AddCandle(candle)

		if store.Len() < 35 {
			continue
		}

		snap := indicator.Build(store.Candles(), cfg.Symbol, cfg.Exchange)
		cls := classifier.Classify(ctx, snap)

		if cls.Confidence < cfg.ConfidenceThresh {
			continue
		}

		allConf = append(allConf, cls.Confidence)

		if cls.Trend != currentTrend {
			if currentTrend != trend.Unknown {
				report.Periods = append(report.Periods, Period{
					Trend:      currentTrend,
					Start:      trendStart,
					End:        candle.Timestamp,
					Duration:   candle.Timestamp.Sub(trendStart),
					CandleSpan: i,
				})
			}
			currentTrend = cls.Trend
			trendStart = candle.Timestamp
			report.ShiftCount++
		}
	}

	// close the final open period
	if currentTrend != trend.Unknown && len(candles) > 0 {
		last := candles[len(candles)-1]
		report.Periods = append(report.Periods, Period{
			Trend:    currentTrend,
			Start:    trendStart,
			End:      last.Timestamp,
			Duration: last.Timestamp.Sub(trendStart),
		})
	}

	report.AvgConfidence = avg(allConf)
	report.TotalCandles = len(candles)

	log.Printf("backtest: processed %d candles, %d trend periods, %d shifts",
		len(candles), len(report.Periods), report.ShiftCount)

	return report
}

// FetchAndRun downloads historical data from Kite and runs the backtest.
func FetchAndRun(ctx context.Context, apiKey, accessToken string, instrumentToken uint32, classifier *trend.Classifier, cfg Config) (Report, error) {
	log.Printf("backtest: fetching %s from %s to %s", cfg.Symbol, cfg.From.Format("2006-01-02"), cfg.To.Format("2006-01-02"))

	candles, err := market.FetchHistorical(apiKey, accessToken, instrumentToken, cfg.From, cfg.To, cfg.KiteInterval)
	if err != nil {
		return Report{}, fmt.Errorf("backtest fetch: %w", err)
	}
	log.Printf("backtest: fetched %d candles", len(candles))

	return Run(ctx, candles, classifier, cfg), nil
}

func avg(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}
