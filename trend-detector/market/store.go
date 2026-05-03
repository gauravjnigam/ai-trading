package market

import (
	"sync"
	"time"
)

// Store holds a fixed-size ring buffer of completed OHLCV candles and builds
// them incrementally from incoming ticks.
type Store struct {
	mu       sync.RWMutex
	candles  []OHLCV
	size     int
	current  *OHLCV
	interval time.Duration

	// partial candle tracking
	candleStart time.Time
}

// NewStore creates a Store that aggregates ticks into candles of the given
// interval and keeps at most size completed candles.
func NewStore(size int, interval time.Duration) *Store {
	return &Store{
		candles:  make([]OHLCV, 0, size),
		size:     size,
		interval: interval,
	}
}

// AddTick incorporates a tick into the current in-progress candle, finalising
// it when the candle interval rolls over.
func (s *Store) AddTick(t Tick) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := t.Timestamp
	bucket := now.Truncate(s.interval)

	if s.current == nil || !s.candleStart.Equal(bucket) {
		if s.current != nil {
			s.append(*s.current)
		}
		s.current = &OHLCV{
			Open:      t.LastPrice,
			High:      t.LastPrice,
			Low:       t.LastPrice,
			Close:     t.LastPrice,
			Volume:    float64(t.Volume),
			Timestamp: bucket,
		}
		s.candleStart = bucket
		return
	}

	if t.LastPrice > s.current.High {
		s.current.High = t.LastPrice
	}
	if t.LastPrice < s.current.Low {
		s.current.Low = t.LastPrice
	}
	s.current.Close = t.LastPrice
	s.current.Volume += float64(t.Volume)
}

// Candles returns a copy of completed candles from oldest to newest.
func (s *Store) Candles() []OHLCV {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]OHLCV, len(s.candles))
	copy(out, s.candles)
	return out
}

// Len returns the number of completed candles stored.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.candles)
}

// AddCandle directly inserts a completed candle (used by the backtester).
func (s *Store) AddCandle(c OHLCV) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.append(c)
}

func (s *Store) append(c OHLCV) {
	if len(s.candles) >= s.size {
		copy(s.candles, s.candles[1:])
		s.candles[s.size-1] = c
	} else {
		s.candles = append(s.candles, c)
	}
}
