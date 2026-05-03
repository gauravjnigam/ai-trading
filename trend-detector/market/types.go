package market

import "time"

// Tick is a single price update from the WebSocket feed.
type Tick struct {
	InstrumentToken uint32
	Symbol          string
	LastPrice       float64
	Volume          uint32
	Timestamp       time.Time
}

// OHLCV is one completed candle.
type OHLCV struct {
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	Timestamp time.Time
}
