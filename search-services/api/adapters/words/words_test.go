package words_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"yadro.com/course/api/adapters/words"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	mockwords "yadro.com/course/api/adapters/words/mock"
	"yadro.com/course/api/core"
	wordspb "yadro.com/course/proto/words"
)

func TestNewClient(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		client, err := words.NewClient("localhost:50051", slog.Default())
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestClient_Norm(t *testing.T) {
	tests := []struct {
		name         string
		phrase       string
		mockResponse *wordspb.WordsReply
		mockError    error
		expected     []string
		expectedErr  error
	}{
		{
			name:   "success",
			phrase: "test phrase",
			mockResponse: &wordspb.WordsReply{
				Words: []string{"test", "phrase"},
			},
			expected: []string{"test", "phrase"},
		},
		{
			name:        "resource exhausted",
			phrase:      "very long phrase",
			mockError:   status.Error(codes.ResourceExhausted, "quota exceeded"),
			expectedErr: core.ErrBadArguments,
		},
		{
			name:        "other error",
			phrase:      "error case",
			mockError:   errors.New("network error"),
			expectedErr: errors.New("network error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mockwords.NewMockWordsClient(ctrl)
			mockClient.EXPECT().
				Norm(gomock.Any(), &wordspb.WordsRequest{Phrase: tt.phrase}, gomock.Any()).
				Return(tt.mockResponse, tt.mockError)

			c := &words.Client{
				Client: mockClient,
				Log:    slog.Default(),
			}

			result, err := c.Norm(context.Background(), tt.phrase)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestClient_Ping(t *testing.T) {
	tests := []struct {
		name        string
		mockError   error
		expectedErr error
	}{
		{
			name: "success",
		},
		{
			name:        "error",
			mockError:   errors.New("ping failed"),
			expectedErr: errors.New("ping failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mockwords.NewMockWordsClient(ctrl)
			mockClient.EXPECT().
				Ping(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
				Return(&emptypb.Empty{}, tt.mockError)

			c := &words.Client{
				Client: mockClient,
				Log:    slog.Default(),
			}

			err := c.Ping(context.Background())

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
