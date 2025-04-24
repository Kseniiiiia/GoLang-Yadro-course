package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/ilyakaznacheev/cleanenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	yolopb "yadro.com/course/proto/yolo"
)

type Config struct {
	Port string `yaml:"yolo_address" env:"YOLO_ADDRESS" env-default:":28085"`
	Yolo string `yaml:"yolo_api_address" env:"YOLO_API_ADDRESS" env-default:":10004"`
}

type server struct {
	yolopb.UnimplementedYoloServiceServer
	log  *slog.Logger
	port string
	url  string
}

func (s *server) Detect(ctx context.Context, req *yolopb.DetectRequest) (*yolopb.DetectResponse, error) {
	binaryImage := req.ImageData

	data := map[string]interface{}{
		"image": map[string]interface{}{
			"py/b64": base64.StdEncoding.EncodeToString(binaryImage),
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	var tmp interface{}
	if err := json.Unmarshal(jsonData, &tmp); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	finalData, err := json.Marshal(tmp)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	escapedJsonString := strconv.Quote(string(finalData))

	resp, err := http.Post(
		s.url,
		"application/json",
		bytes.NewBuffer([]byte(escapedJsonString)),
	)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response error: %w", err)
	}

	var responseDecoded map[string]interface{}
	if err := json.Unmarshal(body, &responseDecoded); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	var response struct {
		YoloResults []struct {
			BBox     []float64 `json:"bbox"`
			DetScore float64   `json:"det_score"`
			LabelNum int       `json:"label_num"`
			Label    string    `json:"label_string"`
		} `json:"yolo_results"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	results := make([]*yolopb.Detection, len(response.YoloResults))
	for i, r := range response.YoloResults {

		bbox := make([]float32, len(r.BBox))
		for j, val := range r.BBox {
			bbox[j] = float32(val)
		}

		results[i] = &yolopb.Detection{
			Bboxes:     bbox,
			Confidence: float32(r.DetScore),
			Label:      r.Label,
			LabelNum:   int32(r.LabelNum),
		}
	}

	return &yolopb.DetectResponse{Results: results}, nil
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "path to config file")
	flag.Parse()

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	if err := run(cfg, log); err != nil {
		log.Error("failed to run server", "error", err)
		os.Exit(1)
	}
}

func run(cfg Config, log *slog.Logger) error {
	lis, err := net.Listen("tcp", cfg.Port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	yolopb.RegisterYoloServiceServer(srv, &server{
		log:  log,
		port: cfg.Port,
		url:  cfg.Yolo,
	})
	reflection.Register(srv)

	log.Info("starting server", "address", cfg.Port)

	return srv.Serve(lis)
}
