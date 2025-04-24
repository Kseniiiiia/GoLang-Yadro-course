package words

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	wordspb "yadro.com/course/proto/words"
	mockwords "yadro.com/course/search/adapters/words/mock"
)

func TestClient_Norm(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockwords.NewMockWordsClient(ctrl)

	tests := []struct {
		name         string
		mockSetup    func()
		phrase       string
		expected     []string
		expectedErr  error
		expectedCode codes.Code
	}{
		{
			name: "successful normalization",
			mockSetup: func() {
				mockClient.EXPECT().Norm(
					gomock.Any(),
					&wordspb.WordsRequest{Phrase: "test phrase"},
				).Return(&wordspb.WordsReply{Words: []string{"test", "phrase"}}, nil)
			},
			phrase:      "test phrase",
			expected:    []string{"test", "phrase"},
			expectedErr: nil,
		},
		{
			name: "empty result",
			mockSetup: func() {
				mockClient.EXPECT().Norm(
					gomock.Any(),
					&wordspb.WordsRequest{Phrase: "empty"},
				).Return(&wordspb.WordsReply{Words: []string{}}, nil)
			},
			phrase:      "empty",
			expected:    []string{},
			expectedErr: nil,
		},
		{
			name: "gRPC error",
			mockSetup: func() {
				mockClient.EXPECT().Norm(
					gomock.Any(),
					&wordspb.WordsRequest{Phrase: "error"},
				).Return(nil, status.Error(codes.Internal, "internal error"))
			},
			phrase:       "error",
			expectedErr:  errors.New("failed to normalize words"),
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			client := &Client{
				client: mockClient,
				log:    slog.Default(),
			}

			result, err := client.Norm(context.Background(), tt.phrase)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())

				if tt.expectedCode != codes.OK {
					st, ok := status.FromError(err)
					assert.True(t, ok)
					assert.Equal(t, tt.expectedCode, st.Code())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestClient_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockwords.NewMockWordsClient(ctrl)

	tests := []struct {
		name         string
		mockSetup    func()
		expectedErr  error
		expectedCode codes.Code
	}{
		{
			name: "successful ping",
			mockSetup: func() {
				mockClient.EXPECT().Ping(
					gomock.Any(),
					gomock.Any(),
				).Return(nil, nil)
			},
			expectedErr: nil,
		},
		{
			name: "ping error",
			mockSetup: func() {
				mockClient.EXPECT().Ping(
					gomock.Any(),
					gomock.Any(),
				).Return(nil, errors.New("ping failed"))
			},
			expectedErr: errors.New("ping failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			client := &Client{
				client: mockClient,
				log:    slog.Default(),
			}

			err := client.Ping(context.Background())

			if tt.expectedErr != nil {
				assert.Error(t, err)

				if tt.expectedCode != codes.OK {
					st, ok := status.FromError(err)
					assert.True(t, ok)
					assert.Equal(t, tt.expectedCode, st.Code())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
