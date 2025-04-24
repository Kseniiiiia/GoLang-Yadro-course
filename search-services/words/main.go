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
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
	"yadro.com/course/words/words"
)

const (
	maxPhraseLen    = 20480
	maxShutdownTime = 5 * time.Second
)

type Config struct {
	Port string `yaml:"words_address" env:"WORDS_ADDRESS" env-default:"8080"`
}

type server struct {
	wordspb.UnimplementedWordsServer
	log *slog.Logger
}

func (s *server) Ping(_ context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *server) Norm(_ context.Context, in *wordspb.WordsRequest) (*wordspb.WordsReply, error) {
	s.log.Debug("norm request", "phrase", in.Phrase)

	if len(in.GetPhrase()) > maxPhraseLen {
		return nil, status.Error(codes.ResourceExhausted, "too large")
	}

	return &wordspb.WordsReply{
		Words: words.Norm(in.GetPhrase()),
	}, nil
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "path to config file")
	flag.Parse()

	var config Config
	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		panic(err)
	}

	log := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
			},
		),
	)

	if err := run(config, log); err != nil {
		log.Error("failed to run server", "error", err)
		os.Exit(1)
	}
}

func run(cfg Config, log *slog.Logger) error {
	listener, err := net.Listen("tcp", cfg.Port)
	if err != nil {
		return fmt.Errorf("failed to listen port %s: %v", cfg.Port, err)
	}

	s := grpc.NewServer()
	wordspb.RegisterWordsServer(s, &server{log: log})
	reflection.Register(s)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("starting server", "port", cfg.Port)
		if err = s.Serve(listener); err != nil {
			log.Error("faied to serve", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	timer := time.AfterFunc(maxShutdownTime, func() {
		log.Info("forcing server stop")
		s.Stop()
	})
	defer timer.Stop()

	log.Info("starting graceful stop")
	s.GracefulStop()
	log.Info("server stopped")

	return err
}
