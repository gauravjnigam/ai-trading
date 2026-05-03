package indicator

import "gnai/trend-detector/market"

// OBVResult holds the current On-Balance Volume and its trend direction.
type OBVResult struct {
	Value   float64
	Rising  bool // true if OBV trended up over the lookback window
}

// OBV computes On-Balance Volume and determines whether it is rising over the
// last lookback candles.
func OBV(candles []market.OHLCV, lookback int) OBVResult {
	if len(candles) < 2 {
		return OBVResult{}
	}

	obv := make([]float64, len(candles))
	for i := 1; i < len(candles); i++ {
		switch {
		case candles[i].Close > candles[i-1].Close:
			obv[i] = obv[i-1] + candles[i].Volume
		case candles[i].Close < candles[i-1].Close:
			obv[i] = obv[i-1] - candles[i].Volume
		default:
			obv[i] = obv[i-1]
		}
	}

	current := obv[len(obv)-1]
	rising := false
	if len(obv) >= lookback+1 {
		rising = current > obv[len(obv)-lookback-1]
	}
	return OBVResult{Value: current, Rising: rising}
}

// VolumeSpike returns true if the latest candle's volume is above the average
// volume of the previous period candles by the given multiplier.
func VolumeSpike(candles []market.OHLCV, period int, multiplier float64) bool {
	if len(candles) < period+1 {
		return false
	}
	window := candles[len(candles)-period-1 : len(candles)-1]
	sum := 0.0
	for _, c := range window {
		sum += c.Volume
	}
	avg := sum / float64(period)
	return candles[len(candles)-1].Volume > avg*multiplier
}
