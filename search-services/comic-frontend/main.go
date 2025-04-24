package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"yadro.com/course/comic-frontend/config"
	"yadro.com/course/comic-frontend/handlers"
)

func main() {
	cfg := config.MustLoad("config.yaml")

	log := setupLogger(cfg.LogLevel)

	client := &http.Client{
		Timeout: cfg.Server.Timeout,
	}

	handler := handlers.NewHandler(log, client, cfg.API.ApiURL)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", handler.Index)
	mux.HandleFunc("GET /search", handler.Search)
	mux.HandleFunc("GET /image-search", handler.ImageSearch)
	mux.HandleFunc("POST /detect", handler.Detect)

	// Админские маршруты
	mux.HandleFunc("GET /admin", handler.Admin)
	mux.HandleFunc("GET /admin/login", handler.AdminLogin)
	mux.HandleFunc("POST /admin/login", handler.AdminLoginPost)
	mux.HandleFunc("POST /admin/update", handler.AdminUpdate)
	mux.HandleFunc("POST /admin/drop", handler.AdminDrop)

	srv := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("Starting server", "address", cfg.Server.Address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed", "error", err)
		}
	}()

	<-ctx.Done()
	log.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("Server shutdown error", "error", err)
	}
	log.Info("Server stopped")
}

func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
}
