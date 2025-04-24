package words

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	mockwords "yadro.com/course/api/adapters/words/mock"
	wordspb "yadro.com/course/proto/words"
)

func TestNewClient(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		client, err := NewClient("localhost:50051", nil)
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestClient_Norm(t *testing.T) {
	tests := []struct {
		name        string
		phrase      string
		mockSetup   func(*mockwords.MockWordsClient)
		expected    []string
		expectedErr string
	}{
		{
			name:   "successful normalization",
			phrase: "test phrase",
			mockSetup: func(m *mockwords.MockWordsClient) {
				m.EXPECT().Norm(
					gomock.Any(),
					&wordspb.WordsRequest{Phrase: "test phrase"},
				).Return(&wordspb.WordsReply{Words: []string{"test", "phrase"}}, nil)
			},
			expected: []string{"test", "phrase"},
		},
		{
			name:   "empty phrase",
			phrase: "",
			mockSetup: func(m *mockwords.MockWordsClient) {
				m.EXPECT().Norm(
					gomock.Any(),
					&wordspb.WordsRequest{Phrase: ""},
				).Return(&wordspb.WordsReply{Words: []string{}}, nil)
			},
			expected: []string{},
		},
		{
			name:   "gRPC error",
			phrase: "error",
			mockSetup: func(m *mockwords.MockWordsClient) {
				m.EXPECT().Norm(
					gomock.Any(),
					&wordspb.WordsRequest{Phrase: "error"},
				).Return(nil, errors.New("gRPC error"))
			},
			expectedErr: "failed to normalize words: gRPC error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mockwords.NewMockWordsClient(ctrl)
			tt.mockSetup(mockClient)

			c := &Client{
				client: mockClient,
				log:    nil,
			}

			result, err := c.Norm(context.Background(), tt.phrase)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestClient_Ping(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*mockwords.MockWordsClient)
		expectedErr error
	}{
		{
			name: "successful ping",
			mockSetup: func(m *mockwords.MockWordsClient) {
				m.EXPECT().Ping(
					gomock.Any(),
					gomock.Any(),
				).Return(nil, nil)
			},
			expectedErr: nil,
		},
		{
			name: "ping error",
			mockSetup: func(m *mockwords.MockWordsClient) {
				m.EXPECT().Ping(
					gomock.Any(),
					gomock.Any(),
				).Return(nil, errors.New("ping failed"))
			},
			expectedErr: errors.New("ping failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mockwords.NewMockWordsClient(ctrl)
			tt.mockSetup(mockClient)

			c := &Client{
				client: mockClient,
				log:    nil,
			}

			err := c.Ping(context.Background())
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
