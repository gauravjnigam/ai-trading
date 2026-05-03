package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gnai/trend-detector/backtest"
	"gnai/trend-detector/market"
	"gnai/trend-detector/notify"
	"gnai/trend-detector/trend"
)

func main() {
	doBacktest := flag.Bool("backtest", false, "run in backtest mode")
	fromFlag := flag.String("from", "", "backtest start date YYYY-MM-DD")
	toFlag := flag.String("to", "", "backtest end date YYYY-MM-DD")
	flag.Parse()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if cfg.AnthropicAPIKey == "" {
		log.Println("ANTHROPIC_API_KEY not set — running in rule-based classifier mode")
	} else {
		log.Println("ANTHROPIC_API_KEY present — running in Claude-enhanced classifier mode")
	}

	classifier := trend.NewClassifier(cfg.AnthropicAPIKey)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if *doBacktest {
		runBacktest(ctx, cfg, classifier, *fromFlag, *toFlag)
		return
	}

	runLive(ctx, cfg, classifier)
}

func runLive(ctx context.Context, cfg config, classifier *trend.Classifier) {
	store := market.NewStore(cfg.StoreLookback, 5*time.Minute)

	streamCfg := market.StreamConfig{
		APIKey:          cfg.KiteAPIKey,
		AccessToken:     cfg.KiteAccessToken,
		InstrumentToken: cfg.KiteInstrumentToken,
		Symbol:          cfg.Symbol,
	}
	go market.StartStream(ctx, streamCfg, store)

	mgr := trend.NewManager(classifier, store, trend.ManagerConfig{
		Symbol:           cfg.Symbol,
		Exchange:         cfg.Exchange,
		CheckInterval:    cfg.TrendCheckInterval,
		DigestInterval:   cfg.DigestInterval,
		Hysteresis:       cfg.Hysteresis,
		ConfidenceThresh: cfg.ConfidenceThresh,
	})

	slack := notify.NewSlack(cfg.SlackWebhookURL, cfg.SlackChannel)
	telegram := notify.NewTelegram(cfg.TelegramBotToken, cfg.TelegramChatID)
	dispatcher := notify.NewDispatcher(slack, telegram)

	go dispatcher.Run(ctx, mgr)

	log.Printf("trend-detector: live mode started for %s:%s", cfg.Exchange, cfg.Symbol)
	mgr.Run(ctx)
	log.Println("trend-detector: shutting down")
}

func runBacktest(ctx context.Context, cfg config, classifier *trend.Classifier, fromFlag, toFlag string) {
	from, to := cfg.BacktestFrom, cfg.BacktestTo

	if fromFlag != "" {
		t, err := time.Parse("2006-01-02", fromFlag)
		if err != nil {
			log.Fatalf("--from: %v", err)
		}
		from = t
	}
	if toFlag != "" {
		t, err := time.Parse("2006-01-02", toFlag)
		if err != nil {
			log.Fatalf("--to: %v", err)
		}
		to = t
	}

	if from.IsZero() || to.IsZero() {
		log.Fatal("backtest requires --from and --to dates (YYYY-MM-DD)")
	}

	btCfg := backtest.Config{
		Symbol:           cfg.Symbol,
		Exchange:         cfg.Exchange,
		From:             from,
		To:               to,
		KiteInterval:     cfg.KiteInterval,
		StoreLookback:    cfg.StoreLookback,
		ConfidenceThresh: cfg.ConfidenceThresh,
	}

	report, err := backtest.FetchAndRun(ctx, cfg.KiteAPIKey, cfg.KiteAccessToken, cfg.KiteInstrumentToken, classifier, btCfg)
	if err != nil {
		log.Fatalf("backtest: %v", err)
	}

	fmt.Println(report.Print())
}
