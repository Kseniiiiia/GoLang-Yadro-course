package search

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	mocksearch "yadro.com/course/api/adapters/search/mock"
	"yadro.com/course/api/core"
	searchpb "yadro.com/course/proto/search"
)

func TestClient_Ping_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocksearch.NewMockSearchClient(ctrl)
	client := &Client{
		client: mockClient,
		log:    slog.Default(),
	}

	mockClient.EXPECT().
		Ping(gomock.Any(), &emptypb.Empty{}).
		Return(&emptypb.Empty{}, nil)

	err := client.Ping(context.Background())
	assert.NoError(t, err)
}

func TestClient_Ping_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocksearch.NewMockSearchClient(ctrl)
	client := &Client{
		client: mockClient,
		log:    slog.Default(),
	}

	expectedErr := status.Error(codes.Internal, "server error")
	mockClient.EXPECT().
		Ping(gomock.Any(), &emptypb.Empty{}).
		Return(nil, expectedErr)

	err := client.Ping(context.Background())
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestClient_Search_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocksearch.NewMockSearchClient(ctrl)
	client := &Client{
		client: mockClient,
		log:    slog.Default(),
	}

	req := &searchpb.SearchRequest{
		Phrase: "xkcd",
		Limit:  10,
	}
	resp := &searchpb.SearchResponse{
		Comics: []*searchpb.Comic{
			{Id: 1, Url: "http://example.com/1"},
			{Id: 2, Url: "http://example.com/2"},
		},
		Total: 2,
	}

	mockClient.EXPECT().
		Search(gomock.Any(), req).
		Return(resp, nil)

	comics, total, err := client.Search(context.Background(), "xkcd", 10)
	assert.NoError(t, err)
	assert.Equal(t, int32(2), total)
	assert.Len(t, comics, 2)
	assert.Equal(t, 1, comics[0].ID)
	assert.Equal(t, "http://example.com/1", comics[0].URL)
}

func TestClient_Search_InvalidArgument(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocksearch.NewMockSearchClient(ctrl)
	client := &Client{
		client: mockClient,
		log:    slog.Default(),
	}

	mockClient.EXPECT().
		Search(gomock.Any(), gomock.Any()).
		Return(nil, status.Error(codes.InvalidArgument, "bad phrase"))

	_, _, err := client.Search(context.Background(), "", 10)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, core.ErrBadArguments))
}

func TestClient_IndexSearch_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocksearch.NewMockSearchClient(ctrl)
	client := &Client{
		client: mockClient,
		log:    slog.Default(),
	}

	req := &searchpb.IndexSearchRequest{
		Phrase: "xkcd",
		Limit:  5,
	}
	resp := &searchpb.SearchResponse{
		Comics: []*searchpb.Comic{
			{Id: 3, Url: "http://example.com/3"},
		},
		Total: 1,
	}

	mockClient.EXPECT().
		IndexSearch(gomock.Any(), req).
		Return(resp, nil)

	comics, total, err := client.IndexSearch(context.Background(), "xkcd", 5)
	assert.NoError(t, err)
	assert.Equal(t, int32(1), total)
	assert.Len(t, comics, 1)
	assert.Equal(t, 3, comics[0].ID)
}

func TestClient_IndexSearch_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocksearch.NewMockSearchClient(ctrl)
	client := &Client{
		client: mockClient,
		log:    slog.Default(),
	}

	mockClient.EXPECT().
		IndexSearch(gomock.Any(), gomock.Any()).
		Return(nil, status.Error(codes.Internal, "indexing failed"))

	_, _, err := client.IndexSearch(context.Background(), "xkcd", 5)
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}
