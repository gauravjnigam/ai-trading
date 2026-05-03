package indicator

import (
	"fmt"
	"time"

	"gnai/trend-detector/market"
)

// Snapshot holds computed values for all indicators at a point in time.
type Snapshot struct {
	Symbol    string
	Exchange  string
	Timestamp time.Time

	// Price
	EMA9        float64
	EMA21       float64
	Supertrend  SupertrendResult
	ADX         ADXResult

	// Momentum
	RSI         float64
	MACD        MACDResult
	Stochastic  StochasticResult

	// Volume
	OBV         OBVResult
	VolumeSpike bool
}

// Build computes all indicators from the candle store and returns a Snapshot.
func Build(candles []market.OHLCV, symbol, exchange string) Snapshot {
	ts := time.Now()
	if len(candles) > 0 {
		ts = candles[len(candles)-1].Timestamp
	}
	return Snapshot{
		Symbol:      symbol,
		Exchange:    exchange,
		Timestamp:   ts,
		EMA9:        EMA(candles, 9),
		EMA21:       EMA(candles, 21),
		Supertrend:  Supertrend(candles, 7, 3),
		ADX:         ADX(candles, 14),
		RSI:         RSI(candles, 14),
		MACD:        MACD(candles),
		Stochastic:  Stochastic(candles, 14, 3),
		OBV:         OBV(candles, 10),
		VolumeSpike: VolumeSpike(candles, 20, 2.0),
	}
}

// Text renders the snapshot as a structured prompt block for the Claude classifier.
func (s Snapshot) Text() string {
	emaCross := "bearish alignment"
	if s.EMA9 > s.EMA21 {
		emaCross = "bullish alignment"
	}

	stDir := "bearish (price below Supertrend)"
	if s.Supertrend.Bullish {
		stDir = "bullish (price above Supertrend)"
	}

	adxStrength := "weak trend"
	switch {
	case s.ADX.ADX >= 40:
		adxStrength = "very strong trend"
	case s.ADX.ADX >= 25:
		adxStrength = "strong trend"
	}

	rsiNote := "neutral"
	switch {
	case s.RSI >= 70:
		rsiNote = "overbought"
	case s.RSI <= 30:
		rsiNote = "oversold"
	}

	macdDir := "flat"
	switch {
	case s.MACD.Histogram > 0:
		macdDir = "rising (bullish)"
	case s.MACD.Histogram < 0:
		macdDir = "falling (bearish)"
	}

	obvDir := "falling"
	if s.OBV.Rising {
		obvDir = "rising (buying pressure)"
	}

	spikeStr := "none"
	if s.VolumeSpike {
		spikeStr = "YES — unusual volume"
	}

	return fmt.Sprintf(`Symbol: %s | Exchange: %s | Time: %s
Timeframe: 5-min candles | Lookback: 50 candles

--- Price Indicators ---
EMA(9): %.2f  |  EMA(21): %.2f  →  %s
Supertrend(7,3): %.2f  →  %s
ADX: %.1f  →  %s  |  +DI: %.1f  -DI: %.1f

--- Momentum Indicators ---
RSI(14): %.1f  →  %s
MACD: %.2f  |  Signal: %.2f  |  Histogram: %.2f  →  %s
Stochastic: %%K %.1f  |  %%D %.1f

--- Volume Indicators ---
OBV (10-candle trend): %s
Volume spike (2× avg): %s

Classify the current market trend.
Choose exactly one: BULLISH | BEARISH | SIDEWAYS | HIGH_VOLATILITY
Return JSON only: {"trend":"...","confidence":0.0,"reasoning":"..."}`,
		s.Symbol, s.Exchange, s.Timestamp.Format("2006-01-02 15:04 MST"),
		s.EMA9, s.EMA21, emaCross,
		s.Supertrend.Value, stDir,
		s.ADX.ADX, adxStrength, s.ADX.PlusDI, s.ADX.MinusDI,
		s.RSI, rsiNote,
		s.MACD.MACD, s.MACD.Signal, s.MACD.Histogram, macdDir,
		s.Stochastic.K, s.Stochastic.D,
		obvDir,
		spikeStr,
	)
}

// RuleBasedVote returns a trend vote from each indicator with its weight,
// used by the rule-based (no-API) classifier mode.
type Vote struct {
	Trend      string
	Weight     float64
	Indicator  string
}

// Votes derives individual indicator votes from the snapshot.
func (s Snapshot) Votes() []Vote {
	var votes []Vote

	// Supertrend (weight 0.25)
	if s.Supertrend.Bullish {
		votes = append(votes, Vote{"BULLISH", 0.25, "Supertrend"})
	} else {
		votes = append(votes, Vote{"BEARISH", 0.25, "Supertrend"})
	}

	// ADX (weight 0.20) — detects high volatility / strong trend
	switch {
	case s.ADX.ADX >= 40:
		votes = append(votes, Vote{"HIGH_VOLATILITY", 0.20, "ADX"})
	case s.ADX.ADX >= 25 && s.ADX.PlusDI > s.ADX.MinusDI:
		votes = append(votes, Vote{"BULLISH", 0.20, "ADX"})
	case s.ADX.ADX >= 25 && s.ADX.MinusDI > s.ADX.PlusDI:
		votes = append(votes, Vote{"BEARISH", 0.20, "ADX"})
	default:
		votes = append(votes, Vote{"SIDEWAYS", 0.20, "ADX"})
	}

	// MACD (weight 0.20)
	switch {
	case s.MACD.Histogram > 0:
		votes = append(votes, Vote{"BULLISH", 0.20, "MACD"})
	case s.MACD.Histogram < 0:
		votes = append(votes, Vote{"BEARISH", 0.20, "MACD"})
	default:
		votes = append(votes, Vote{"SIDEWAYS", 0.20, "MACD"})
	}

	// RSI (weight 0.15)
	switch {
	case s.RSI >= 60:
		votes = append(votes, Vote{"BULLISH", 0.15, "RSI"})
	case s.RSI <= 40:
		votes = append(votes, Vote{"BEARISH", 0.15, "RSI"})
	default:
		votes = append(votes, Vote{"SIDEWAYS", 0.15, "RSI"})
	}

	// EMA crossover (weight 0.12)
	if s.EMA9 > s.EMA21 {
		votes = append(votes, Vote{"BULLISH", 0.12, "EMA crossover"})
	} else {
		votes = append(votes, Vote{"BEARISH", 0.12, "EMA crossover"})
	}

	// OBV (weight 0.08)
	if s.OBV.Rising {
		votes = append(votes, Vote{"BULLISH", 0.08, "OBV"})
	} else {
		votes = append(votes, Vote{"BEARISH", 0.08, "OBV"})
	}

	return votes
}
