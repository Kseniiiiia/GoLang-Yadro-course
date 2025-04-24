package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	searchpb "yadro.com/course/proto/search"
	"yadro.com/course/search/adapters/db"
	searchgrpc "yadro.com/course/search/adapters/grpc"
	"yadro.com/course/search/adapters/initiator"
	"yadro.com/course/search/adapters/words"
	"yadro.com/course/search/config"
	"yadro.com/course/search/core"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "path to config file")
	flag.Parse()

	cfg := config.MustLoad(configPath)

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	if err := run(cfg, log); err != nil {
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, log *slog.Logger) error {
	dbAdapter, err := db.New(log, cfg.DBAddress)
	if err != nil {
		log.Error("failed to connect to db", "error", err)
		os.Exit(1)
	}

	wordsAdapter, err := words.NewClient(cfg.WordsAddress, log)
	if err != nil {
		log.Error("failed to connect to words service", "error", err)
		os.Exit(1)
	}

	service, err := core.NewService(log, dbAdapter, wordsAdapter)
	if err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	go func() {
		init := initiator.NewInit(log, service, cfg.SearchConfig.IndexTTL)
		init.Start(context.Background())
	}()

	listener, err := net.Listen("tcp", cfg.SearchConfig.Address)
	if err != nil {
		log.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	s := grpc.NewServer()
	searchpb.RegisterSearchServer(s, searchgrpc.NewServer(service))
	reflection.Register(s)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Info("shutting down server")
		s.GracefulStop()
	}()

	log.Info("starting server", "address", cfg.SearchConfig.Address)

	if err := s.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}
