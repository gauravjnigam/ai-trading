package indicator

import "gnai/trend-detector/market"

// RSI computes the Relative Strength Index for the given period.
func RSI(candles []market.OHLCV, period int) float64 {
	if len(candles) < period+1 {
		return 0
	}

	gains, losses := 0.0, 0.0
	for i := 1; i <= period; i++ {
		diff := candles[i].Close - candles[i-1].Close
		if diff > 0 {
			gains += diff
		} else {
			losses -= diff
		}
	}
	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	for i := period + 1; i < len(candles); i++ {
		diff := candles[i].Close - candles[i-1].Close
		gain, loss := 0.0, 0.0
		if diff > 0 {
			gain = diff
		} else {
			loss = -diff
		}
		avgGain = (avgGain*float64(period-1) + gain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)
	}

	if avgLoss == 0 {
		return 100
	}
	rs := avgGain / avgLoss
	return 100 - 100/(1+rs)
}

// MACDResult holds the MACD line, signal line, and histogram.
type MACDResult struct {
	MACD      float64
	Signal    float64
	Histogram float64
}

// MACD computes Moving Average Convergence Divergence using the standard
// 12/26/9 configuration.
func MACD(candles []market.OHLCV) MACDResult {
	if len(candles) < 35 {
		return MACDResult{}
	}
	fast := emaSlice(candles, 12)
	slow := emaSlice(candles, 26)

	n := min2(len(fast), len(slow))
	macdLine := make([]float64, n)
	for i := range macdLine {
		macdLine[i] = fast[len(fast)-n+i] - slow[len(slow)-n+i]
	}

	signal := emaOfSlice(macdLine, 9)
	macdVal := macdLine[len(macdLine)-1]
	return MACDResult{
		MACD:      macdVal,
		Signal:    signal,
		Histogram: macdVal - signal,
	}
}

// StochasticResult holds the %K and %D values.
type StochasticResult struct {
	K float64
	D float64
}

// Stochastic computes the Stochastic oscillator (%K and %D).
func Stochastic(candles []market.OHLCV, kPeriod, dPeriod int) StochasticResult {
	if len(candles) < kPeriod+dPeriod {
		return StochasticResult{}
	}

	kValues := make([]float64, len(candles)-kPeriod+1)
	for i := kPeriod - 1; i < len(candles); i++ {
		window := candles[i-kPeriod+1 : i+1]
		lo, hi := window[0].Low, window[0].High
		for _, c := range window[1:] {
			if c.Low < lo {
				lo = c.Low
			}
			if c.High > hi {
				hi = c.High
			}
		}
		if hi == lo {
			kValues[i-kPeriod+1] = 50
		} else {
			kValues[i-kPeriod+1] = 100 * (candles[i].Close - lo) / (hi - lo)
		}
	}

	k := kValues[len(kValues)-1]
	d := 0.0
	if len(kValues) >= dPeriod {
		for _, v := range kValues[len(kValues)-dPeriod:] {
			d += v
		}
		d /= float64(dPeriod)
	}
	return StochasticResult{K: k, D: d}
}

// --- helpers ---

func emaSlice(candles []market.OHLCV, period int) []float64 {
	if len(candles) < period {
		return nil
	}
	k := 2.0 / float64(period+1)
	result := make([]float64, len(candles)-period+1)
	sum := 0.0
	for _, c := range candles[:period] {
		sum += c.Close
	}
	result[0] = sum / float64(period)
	for i, c := range candles[period:] {
		result[i+1] = c.Close*k + result[i]*(1-k)
	}
	return result
}

func emaOfSlice(data []float64, period int) float64 {
	if len(data) < period {
		return 0
	}
	k := 2.0 / float64(period+1)
	sum := 0.0
	for _, v := range data[:period] {
		sum += v
	}
	ema := sum / float64(period)
	for _, v := range data[period:] {
		ema = v*k + ema*(1-k)
	}
	return ema
}

func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}
