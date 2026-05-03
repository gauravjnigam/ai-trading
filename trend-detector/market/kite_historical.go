package market

import (
	"fmt"
	"time"

	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

// FetchHistorical downloads OHLCV candles from the Kite historical data API
// and returns them sorted oldest-first.
func FetchHistorical(apiKey, accessToken string, instrumentToken uint32, from, to time.Time, interval string) ([]OHLCV, error) {
	kc := kiteconnect.New(apiKey)
	kc.SetAccessToken(accessToken)

	records, err := kc.GetHistoricalData(
		int(instrumentToken),
		interval,
		from,
		to,
		false, // continuous
		false, // OI
	)
	if err != nil {
		return nil, fmt.Errorf("kite historical fetch: %w", err)
	}

	candles := make([]OHLCV, 0, len(records))
	for _, r := range records {
		candles = append(candles, OHLCV{
			Open:      r.Open,
			High:      r.High,
			Low:       r.Low,
			Close:     r.Close,
			Volume:    float64(r.Volume),
			Timestamp: r.Date.Time,
		})
	}
	return candles, nil
}
