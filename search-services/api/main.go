package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"yadro.com/course/api/adapters/aaa"
	"yadro.com/course/api/adapters/rest/middleware"
	"yadro.com/course/api/adapters/yolo"

	"yadro.com/course/api/adapters/search"

	"yadro.com/course/api/adapters/words"
	"yadro.com/course/api/core"

	"yadro.com/course/api/adapters/rest"
	"yadro.com/course/api/adapters/update"
	"yadro.com/course/api/config"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()

	cfg := config.MustLoad(configPath)

	log := mustMakeLogger(cfg.LogLevel)

	log.Info("starting server")
	log.Debug("debug messages are enabled")

	tokenTTL := 2 * time.Minute
	if ttl := os.Getenv("TOKEN_TTL"); ttl != "" {
		if parsed, err := time.ParseDuration(ttl); err == nil {
			tokenTTL = parsed
		}
	}

	aaaService, err := aaa.New(tokenTTL, log)
	if err != nil {
		log.Error("cannot init AAA service", "error", err)
		os.Exit(1)
	}

	concurrencyLimit, _ := strconv.Atoi(os.Getenv("SEARCH_CONCURRENCY"))
	rateLimit, _ := strconv.Atoi(os.Getenv("SEARCH_RATE"))

	updateClient, err := update.NewClient(cfg.UpdateAddress, log)
	if err != nil {
		log.Error("cannot init update adapter", "error", err)
		os.Exit(1)
	}

	wordsClient, err := words.NewClient(cfg.WordsAddress, log)
	if err != nil {
		log.Error("cannot init words adapter", "error", err)
		os.Exit(1)
	}

	searchClient, err := search.NewClient(cfg.SearchAddress, log)
	if err != nil {
		log.Error("cannot init search adapter", "error", err)
		os.Exit(1)
	}

	yoloClient, err := yolo.NewClient(cfg.YoloAddress, log)
	if err != nil {
		log.Error("failed to create yolo client", "error", err)
		return
	}

	mux := http.NewServeMux()
	mux.Handle("POST  /api/detect", rest.NewDetectHandler(log, yoloClient, searchClient))

	mux.Handle("POST /api/login", rest.NewLoginHandler(log, aaaService))
	mux.Handle("GET /api/db/stats", rest.NewUpdateStatsHandler(log, updateClient))
	mux.Handle("GET /api/db/status", rest.NewUpdateStatusHandler(log, updateClient))
	mux.Handle("GET /api/words", rest.NewWordsHandler(log, wordsClient))
	mux.Handle("GET /api/ping", rest.NewPingHandler(log, map[string]core.Pinger{"words": wordsClient, "update": updateClient, "search": searchClient}))

	mux.Handle("POST /api/db/update", middleware.Auth(
		rest.NewUpdateHandler(log, updateClient),
		aaaService,
	))

	mux.Handle("DELETE /api/db", middleware.Auth(
		rest.NewDropHandler(log, updateClient),
		aaaService,
	))

	mux.Handle("GET /api/search", middleware.Concurrency(
		rest.NewSearchHandler(log, searchClient),
		concurrencyLimit,
	))

	mux.Handle("GET /api/isearch", middleware.Rate(
		rest.NewSearchIndexHandler(log, searchClient),
		rateLimit,
	))
	server := http.Server{
		Addr:        cfg.HTTPConfig.Address,
		ReadTimeout: cfg.HTTPConfig.Timeout,
		Handler:     mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Error("erroneous shutdown", "error", err)
		}
	}()

	log.Info("Running HTTP server", "address", cfg.HTTPConfig.Address)
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("server closed unexpectedly", "error", err)
			return
		}
	}
}

func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		panic("unknown log level: " + logLevel)
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level, AddSource: true})
	return slog.New(handler)
}
