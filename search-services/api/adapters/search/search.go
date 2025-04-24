package search

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"yadro.com/course/api/core"
	searchpb "yadro.com/course/proto/search"
)

type Client struct {
	log    *slog.Logger
	client searchpb.SearchClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("failed to connect to search service", "error", err)
		return nil, err
	}
	return &Client{
		client: searchpb.NewSearchClient(conn),
		log:    log,
	}, nil
}

func (c Client) Ping(ctx context.Context) error {
	c.log.Debug("calling Ping")
	_, err := c.client.Ping(ctx, &emptypb.Empty{})
	if err != nil {
		c.log.Error("error calling Ping", "error", err)
		return err
	}
	c.log.Debug("successfully pinged search")
	return nil
}

func (c Client) Search(ctx context.Context, phrase string, limit int32) ([]core.Comics, int32, error) {
	c.log.Debug("calling Search", "phrase", phrase, "limit", limit)
	resp, err := c.client.Search(ctx, &searchpb.SearchRequest{
		Phrase: phrase,
		Limit:  limit,
	})
	if err != nil {
		if status.Code(err) == codes.InvalidArgument {
			c.log.Warn("invalid argument", "error", err)
			return nil, 0, core.ErrBadArguments
		}
		c.log.Error("error calling Search", "error", err)
		return nil, 0, err
	}

	var comics []core.Comics
	for _, comic := range resp.Comics {
		comics = append(comics, core.Comics{
			ID:  int(comic.Id),
			URL: comic.Url,
		})
	}

	c.log.Debug("successfully searched comics", "total", resp.Total)
	return comics, resp.Total, nil
}

func (c Client) IndexSearch(ctx context.Context, phrase string, limit int32) ([]core.Comics, int32, error) {
	c.log.Debug("calling IndexSearch", "phrase", phrase, "limit", limit)

	resp, err := c.client.IndexSearch(ctx, &searchpb.IndexSearchRequest{
		Phrase: phrase,
		Limit:  limit,
	})
	if err != nil {
		if status.Code(err) == codes.InvalidArgument {
			c.log.Warn("invalid argument in IndexSearch", "error", err)
			return nil, 0, core.ErrBadArguments
		}
		c.log.Error("error calling IndexSearch", "error", err)
		return nil, 0, err
	}

	var comics []core.Comics
	for _, comic := range resp.Comics {
		comics = append(comics, core.Comics{
			ID:  int(comic.Id),
			URL: comic.Url,
		})
	}

	c.log.Debug("successfully searched comics via index", "total", resp.Total)
	return comics, resp.Total, nil
}
