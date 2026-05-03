package market

import (
	"context"
	"log"
	"time"

	"github.com/zerodha/gokiteconnect/v4/models"
	kiteticker "github.com/zerodha/gokiteconnect/v4/ticker"
)

// StreamConfig holds Kite credentials and the instrument to stream.
type StreamConfig struct {
	APIKey          string
	AccessToken     string
	InstrumentToken uint32
	Symbol          string
}

// StartStream connects to the Kite WebSocket, feeds ticks into store, and
// reconnects automatically until ctx is cancelled.
func StartStream(ctx context.Context, cfg StreamConfig, store *Store) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		runStream(ctx, cfg, store)

		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
			log.Println("kite_stream: reconnecting...")
		}
	}
}

func runStream(ctx context.Context, cfg StreamConfig, store *Store) {
	ticker := kiteticker.New(cfg.APIKey, cfg.AccessToken)
	ticker.SetAutoReconnect(false)

	tokens := []uint32{cfg.InstrumentToken}

	ticker.OnConnect(func() {
		log.Printf("kite_stream: connected, subscribing %d", cfg.InstrumentToken)
		if err := ticker.Subscribe(tokens); err != nil {
			log.Printf("kite_stream: subscribe error: %v", err)
			return
		}
		if err := ticker.SetMode(kiteticker.ModeFull, tokens); err != nil {
			log.Printf("kite_stream: set mode error: %v", err)
		}
	})

	ticker.OnTick(func(tick models.Tick) {
		store.AddTick(Tick{
			InstrumentToken: tick.InstrumentToken,
			Symbol:          cfg.Symbol,
			LastPrice:       tick.LastPrice,
			Volume:          tick.VolumeTraded,
			Timestamp:       tick.Timestamp.Time,
		})
	})

	ticker.OnError(func(err error) {
		log.Printf("kite_stream: error: %v", err)
	})

	ticker.OnClose(func(code int, reason string) {
		log.Printf("kite_stream: closed: %d %s", code, reason)
	})

	ticker.ServeWithContext(ctx)
}
