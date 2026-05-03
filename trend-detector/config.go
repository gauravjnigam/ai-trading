package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type config struct {
	// Kite / Zerodha
	KiteAPIKey         string
	KiteAccessToken    string
	KiteInstrumentToken uint32
	KiteInterval       string

	// Claude (optional)
	AnthropicAPIKey string

	// Symbol
	Symbol   string
	Exchange string

	// Timing
	TrendCheckInterval time.Duration
	DigestInterval     time.Duration
	Hysteresis         int
	ConfidenceThresh   float64
	StoreLookback      int

	// Notifications
	SlackWebhookURL string
	SlackChannel    string
	TelegramBotToken string
	TelegramChatID  string

	// Backtest
	BacktestFrom time.Time
	BacktestTo   time.Time
}

func loadConfig() (config, error) {
	cfg := config{
		KiteAPIKey:         mustEnv("KITE_API_KEY"),
		KiteAccessToken:    mustEnv("KITE_ACCESS_TOKEN"),
		AnthropicAPIKey:    os.Getenv("ANTHROPIC_API_KEY"),
		Symbol:             envOr("TRADING_SYMBOL", "NIFTY 50"),
		Exchange:           envOr("TRADING_EXCHANGE", "NSE"),
		KiteInterval:       envOr("KITE_INTERVAL", "5minute"),
		TrendCheckInterval: envDuration("TREND_CHECK_INTERVAL_SEC", 60) * time.Second,
		DigestInterval:     envDuration("DIGEST_INTERVAL_MIN", 30) * time.Minute,
		Hysteresis:         envInt("SHIFT_HYSTERESIS_COUNT", 3),
		ConfidenceThresh:   envFloat("CONFIDENCE_THRESHOLD", 0.65),
		StoreLookback:      envInt("STORE_LOOKBACK", 100),
		SlackWebhookURL:    os.Getenv("SLACK_WEBHOOK_URL"),
		SlackChannel:       envOr("SLACK_CHANNEL", "#trading-alerts"),
		TelegramBotToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:     os.Getenv("TELEGRAM_CHAT_ID"),
	}

	token, err := strconv.ParseUint(os.Getenv("KITE_INSTRUMENT_TOKEN"), 10, 32)
	if err == nil {
		cfg.KiteInstrumentToken = uint32(token)
	}

	if from := os.Getenv("BACKTEST_FROM"); from != "" {
		cfg.BacktestFrom, err = time.Parse("2006-01-02", from)
		if err != nil {
			return config{}, fmt.Errorf("BACKTEST_FROM: %w", err)
		}
	}
	if to := os.Getenv("BACKTEST_TO"); to != "" {
		cfg.BacktestTo, err = time.Parse("2006-01-02", to)
		if err != nil {
			return config{}, fmt.Errorf("BACKTEST_TO: %w", err)
		}
	}

	return cfg, nil
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fmt.Fprintf(os.Stderr, "required env var %s is not set\n", key)
	}
	return v
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envFloat(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func envDuration(key string, defVal int) time.Duration {
	return time.Duration(envInt(key, defVal))
}
