package yolo

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"yadro.com/course/api/core"
	yolopb "yadro.com/course/proto/yolo"
)

type Client struct {
	client yolopb.YoloServiceClient
	conn   *grpc.ClientConn
	log    *slog.Logger
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("failed to connect to yolo service", "error", err)
		return nil, err
	}
	return &Client{
		client: yolopb.NewYoloServiceClient(conn),
		conn:   conn,
		log:    log,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Detect(ctx context.Context, imageData []byte) ([]core.Yolo, error) {
	resp, err := c.client.Detect(ctx, &yolopb.DetectRequest{
		ImageData: imageData,
	})
	if err != nil {
		return nil, err
	}

	results := make([]core.Yolo, len(resp.Results))
	for i, r := range resp.Results {
		results[i] = core.Yolo{
			BBox:       r.Bboxes,
			Confidence: r.Confidence,
			Label:      r.Label,
			LabelNum:   int(r.LabelNum),
		}
	}
	return results, nil
}
