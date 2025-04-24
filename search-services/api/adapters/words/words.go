package words

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/core"
	wordspb "yadro.com/course/proto/words"
)

type Client struct {
	Log    *slog.Logger
	Client wordspb.WordsClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	connection, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("failed to connect to server", "error", err)
		return nil, err
	}

	return &Client{
		Log:    log,
		Client: wordspb.NewWordsClient(connection),
	}, nil
}

func (c Client) Norm(ctx context.Context, phrase string) ([]string, error) {
	c.Log.Debug("calling Norm", "phrase", phrase)
	resp, err := c.Client.Norm(ctx, &wordspb.WordsRequest{Phrase: phrase})
	if err != nil {
		if status.Code(err) == codes.ResourceExhausted {
			c.Log.Warn("resource exhausted", "error", err)
			return nil, core.ErrBadArguments
		}
		c.Log.Error("error calling Norm", "error", err)
		return nil, err
	}
	c.Log.Debug("successfully normalized phrase", "words", resp.Words)
	return resp.Words, nil
}

func (c Client) Ping(ctx context.Context) error {
	c.Log.Debug("calling Ping")
	_, err := c.Client.Ping(ctx, &emptypb.Empty{})
	if err != nil {
		c.Log.Error("error calling Ping", "error", err)
		return err
	}
	c.Log.Debug("successfully pinged words")
	return nil
}
