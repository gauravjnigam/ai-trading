package indicator

import "gnai/trend-detector/market"

// EMA computes the Exponential Moving Average for the given period.
// Returns 0 if there are fewer candles than the period.
func EMA(candles []market.OHLCV, period int) float64 {
	if len(candles) < period {
		return 0
	}
	k := 2.0 / float64(period+1)
	ema := SMA(candles[:period], period)
	for _, c := range candles[period:] {
		ema = c.Close*k + ema*(1-k)
	}
	return ema
}

// SMA computes the Simple Moving Average over the last n candles.
func SMA(candles []market.OHLCV, period int) float64 {
	n := len(candles)
	if n < period {
		return 0
	}
	sum := 0.0
	for _, c := range candles[n-period:] {
		sum += c.Close
	}
	return sum / float64(period)
}

// ADXResult holds ADX and the directional indicators.
type ADXResult struct {
	ADX float64
	PlusDI  float64
	MinusDI float64
}

// ADX computes the Average Directional Index (14-period standard).
func ADX(candles []market.OHLCV, period int) ADXResult {
	if len(candles) < period*2 {
		return ADXResult{}
	}

	trueRanges := make([]float64, len(candles)-1)
	plusDMs := make([]float64, len(candles)-1)
	minusDMs := make([]float64, len(candles)-1)

	for i := 1; i < len(candles); i++ {
		prev := candles[i-1]
		curr := candles[i]

		hl := curr.High - curr.Low
		hc := abs(curr.High - prev.Close)
		lc := abs(curr.Low - prev.Close)
		trueRanges[i-1] = max3(hl, hc, lc)

		upMove := curr.High - prev.High
		downMove := prev.Low - curr.Low

		if upMove > downMove && upMove > 0 {
			plusDMs[i-1] = upMove
		}
		if downMove > upMove && downMove > 0 {
			minusDMs[i-1] = downMove
		}
	}

	smoothTR := smooth(trueRanges, period)
	smoothPlus := smooth(plusDMs, period)
	smoothMinus := smooth(minusDMs, period)

	if smoothTR == 0 {
		return ADXResult{}
	}

	plusDI := 100 * smoothPlus / smoothTR
	minusDI := 100 * smoothMinus / smoothTR

	dx := 0.0
	if plusDI+minusDI > 0 {
		dx = 100 * abs(plusDI-minusDI) / (plusDI + minusDI)
	}

	return ADXResult{ADX: dx, PlusDI: plusDI, MinusDI: minusDI}
}

// SupertrendResult holds the Supertrend value and direction.
type SupertrendResult struct {
	Value    float64
	Bullish  bool // true = price above supertrend (bullish)
}

// Supertrend computes the Supertrend indicator using ATR.
func Supertrend(candles []market.OHLCV, period int, multiplier float64) SupertrendResult {
	if len(candles) < period+1 {
		return SupertrendResult{}
	}

	atrs := atrSeries(candles, period)
	n := len(candles)

	upperBand := (candles[n-1].High+candles[n-1].Low)/2 + multiplier*atrs[len(atrs)-1]
	lowerBand := (candles[n-1].High+candles[n-1].Low)/2 - multiplier*atrs[len(atrs)-1]

	bullish := candles[n-1].Close > lowerBand

	st := lowerBand
	if !bullish {
		st = upperBand
	}
	return SupertrendResult{Value: st, Bullish: bullish}
}

// --- helpers ---

func atrSeries(candles []market.OHLCV, period int) []float64 {
	trs := make([]float64, len(candles)-1)
	for i := 1; i < len(candles); i++ {
		trs[i-1] = max3(
			candles[i].High-candles[i].Low,
			abs(candles[i].High-candles[i-1].Close),
			abs(candles[i].Low-candles[i-1].Close),
		)
	}
	out := make([]float64, len(trs))
	if len(trs) < period {
		return out
	}
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += trs[i]
	}
	out[period-1] = sum / float64(period)
	for i := period; i < len(trs); i++ {
		out[i] = (out[i-1]*float64(period-1) + trs[i]) / float64(period)
	}
	return out
}

func smooth(data []float64, period int) float64 {
	if len(data) < period {
		return 0
	}
	sum := 0.0
	for _, v := range data[:period] {
		sum += v
	}
	result := sum
	for _, v := range data[period:] {
		result = result - result/float64(period) + v
	}
	return result / float64(period)
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func max3(a, b, c float64) float64 {
	if a >= b && a >= c {
		return a
	}
	if b >= c {
		return b
	}
	return c
}
