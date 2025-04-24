package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	searchpb "yadro.com/course/proto/search"
	"yadro.com/course/search/core"
)

func NewServer(service core.Searcher) *Server {
	return &Server{service: service}
}

type Server struct {
	searchpb.UnimplementedSearchServer
	service core.Searcher
}

func (s *Server) Search(ctx context.Context, req *searchpb.SearchRequest) (*searchpb.SearchResponse, error) {
	result, err := s.service.Search(ctx, req.Phrase, int(req.Limit))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var comics []*searchpb.Comic
	for _, comic := range result.Comics {
		comics = append(comics, &searchpb.Comic{
			Id:  int32(comic.ID),
			Url: comic.URL,
		})
	}

	return &searchpb.SearchResponse{
		Comics: comics,
		Total:  int32(result.Total),
	}, nil
}

func (s *Server) IndexSearch(ctx context.Context, req *searchpb.IndexSearchRequest) (*searchpb.SearchResponse, error) {
	result, err := s.service.IndexSearch(ctx, req.Phrase, int(req.Limit))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var comics []*searchpb.Comic
	for _, comic := range result.Comics {
		comics = append(comics, &searchpb.Comic{
			Id:  int32(comic.ID),
			Url: comic.URL,
		})
	}
	return &searchpb.SearchResponse{
		Comics: comics,
		Total:  int32(result.Total),
	}, nil
}

func (s *Server) Ping(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
