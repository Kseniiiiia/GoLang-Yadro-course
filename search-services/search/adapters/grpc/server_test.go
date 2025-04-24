package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	searchpb "yadro.com/course/proto/search"
	mockserver "yadro.com/course/search/adapters/grpc/mock"
	"yadro.com/course/search/core"
)

func TestNewServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mockserver.NewMockSearcher(ctrl)
	server := NewServer(mockService)

	assert.NotNil(t, server)
	assert.Equal(t, mockService, server.service)
}

func TestServer_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mockserver.NewMockSearcher(ctrl)
	server := NewServer(mockService)

	resp, err := server.Ping(context.Background(), &emptypb.Empty{})

	assert.NoError(t, err)
	assert.IsType(t, &emptypb.Empty{}, resp)
}

func TestServer_Search(t *testing.T) {
	tests := []struct {
		name         string
		mockSetup    func(*mockserver.MockSearcher)
		req          *searchpb.SearchRequest
		expectedResp *searchpb.SearchResponse
		expectedErr  error
		expectedCode codes.Code
	}{
		{
			name: "Successful search",
			mockSetup: func(m *mockserver.MockSearcher) {
				m.EXPECT().Search(gomock.Any(), "test", 10).
					Return(core.SearchResult{
						Comics: []core.Comics{
							{ID: 1, URL: "http://example.com/1"},
							{ID: 2, URL: "http://example.com/2"},
						},
						Total: 2,
					}, nil)
			},
			req: &searchpb.SearchRequest{
				Phrase: "test",
				Limit:  10,
			},
			expectedResp: &searchpb.SearchResponse{
				Comics: []*searchpb.Comic{
					{Id: 1, Url: "http://example.com/1"},
					{Id: 2, Url: "http://example.com/2"},
				},
				Total: 2,
			},
			expectedErr: nil,
		},
		{
			name: "Empty result",
			mockSetup: func(m *mockserver.MockSearcher) {
				m.EXPECT().Search(gomock.Any(), "empty", 10).
					Return(core.SearchResult{
						Comics: []core.Comics{},
						Total:  0,
					}, nil)
			},
			req: &searchpb.SearchRequest{
				Phrase: "empty",
				Limit:  10,
			},
			expectedResp: &searchpb.SearchResponse{
				Comics: []*searchpb.Comic(nil),
				Total:  0,
			},
			expectedErr: nil,
		},
		{
			name: "Internal error",
			mockSetup: func(m *mockserver.MockSearcher) {
				m.EXPECT().Search(gomock.Any(), "error", 10).
					Return(core.SearchResult{}, errors.New("search error"))
			},
			req: &searchpb.SearchRequest{
				Phrase: "error",
				Limit:  10,
			},
			expectedErr:  errors.New("search error"),
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mockserver.NewMockSearcher(ctrl)
			tt.mockSetup(mockService)

			server := NewServer(mockService)
			resp, err := server.Search(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Contains(t, st.Message(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}

func TestServer_IndexSearch(t *testing.T) {
	tests := []struct {
		name         string
		mockSetup    func(*mockserver.MockSearcher)
		req          *searchpb.IndexSearchRequest
		expectedResp *searchpb.SearchResponse
		expectedErr  error
		expectedCode codes.Code
	}{
		{
			name: "Successful index search",
			mockSetup: func(m *mockserver.MockSearcher) {
				m.EXPECT().IndexSearch(gomock.Any(), "test", 5).
					Return(core.SearchResult{
						Comics: []core.Comics{
							{ID: 3, URL: "http://example.com/3"},
						},
						Total: 1,
					}, nil)
			},
			req: &searchpb.IndexSearchRequest{
				Phrase: "test",
				Limit:  5,
			},
			expectedResp: &searchpb.SearchResponse{
				Comics: []*searchpb.Comic{
					{Id: 3, Url: "http://example.com/3"},
				},
				Total: 1,
			},
			expectedErr: nil,
		},
		{
			name: "Error in index search",
			mockSetup: func(m *mockserver.MockSearcher) {
				m.EXPECT().IndexSearch(gomock.Any(), "error", 5).
					Return(core.SearchResult{}, errors.New("index search error"))
			},
			req: &searchpb.IndexSearchRequest{
				Phrase: "error",
				Limit:  5,
			},
			expectedErr:  errors.New("index search error"),
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mockserver.NewMockSearcher(ctrl)
			tt.mockSetup(mockService)

			server := NewServer(mockService)
			resp, err := server.IndexSearch(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Contains(t, st.Message(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}
