package trend

import (
	"context"
	"log"
	"time"

	"gnai/trend-detector/indicator"
	"gnai/trend-detector/market"
)

// Manager continuously classifies trend and emits events on shifts or digests.
type Manager struct {
	classifier        *Classifier
	store             *market.Store
	symbol            string
	exchange          string
	checkInterval     time.Duration
	digestInterval    time.Duration
	hysteresis        int  // consecutive agreements needed to confirm a shift
	confidenceThresh  float64

	// internal state
	current           Classification
	pendingTrend      TrendType
	pendingCount      int
	startedAt         time.Time

	ShiftEvents  chan State
	DigestEvents chan State
}

// ManagerConfig holds all tunable parameters for the Manager.
type ManagerConfig struct {
	Symbol            string
	Exchange          string
	CheckInterval     time.Duration
	DigestInterval    time.Duration
	Hysteresis        int
	ConfidenceThresh  float64
}

// NewManager creates a Manager wired to the given classifier and store.
func NewManager(classifier *Classifier, store *market.Store, cfg ManagerConfig) *Manager {
	return &Manager{
		classifier:       classifier,
		store:            store,
		symbol:           cfg.Symbol,
		exchange:         cfg.Exchange,
		checkInterval:    cfg.CheckInterval,
		digestInterval:   cfg.DigestInterval,
		hysteresis:       cfg.Hysteresis,
		confidenceThresh: cfg.ConfidenceThresh,
		current:          Classification{Trend: Unknown},
		startedAt:        time.Now(),
		ShiftEvents:      make(chan State, 8),
		DigestEvents:     make(chan State, 8),
	}
}

// Run starts the classification loop. It blocks until ctx is cancelled.
func (m *Manager) Run(ctx context.Context) {
	checkTicker := time.NewTicker(m.checkInterval)
	digestTicker := time.NewTicker(m.digestInterval)
	defer checkTicker.Stop()
	defer digestTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-checkTicker.C:
			m.tick(ctx)

		case <-digestTicker.C:
			if m.current.Trend != Unknown {
				m.DigestEvents <- m.buildState(false)
			}
		}
	}
}

func (m *Manager) tick(ctx context.Context) {
	candles := m.store.Candles()
	if len(candles) < 35 {
		log.Printf("state: not enough candles (%d), skipping", len(candles))
		return
	}

	snap := indicator.Build(candles, m.symbol, m.exchange)
	cls := m.classifier.Classify(ctx, snap)

	if cls.Confidence < m.confidenceThresh {
		return
	}

	// hysteresis: only confirm a new trend after N consecutive same readings
	if cls.Trend != m.current.Trend {
		if cls.Trend == m.pendingTrend {
			m.pendingCount++
		} else {
			m.pendingTrend = cls.Trend
			m.pendingCount = 1
		}

		if m.pendingCount >= m.hysteresis {
			prev := m.current
			m.current = cls
			m.startedAt = time.Now()
			m.pendingTrend = Unknown
			m.pendingCount = 0

			st := State{
				Current:   cls,
				Previous:  prev,
				StartedAt: m.startedAt,
				Shifted:   true,
			}
			select {
			case m.ShiftEvents <- st:
			default:
				log.Println("state: shift event channel full, dropping")
			}
		}
	} else {
		// Same trend — reset pending counter
		m.pendingTrend = Unknown
		m.pendingCount = 0
	}
}

func (m *Manager) buildState(shifted bool) State {
	return State{
		Current:   m.current,
		StartedAt: m.startedAt,
		Shifted:   shifted,
	}
}
