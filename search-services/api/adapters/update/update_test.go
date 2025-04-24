package update

import (
	"context"
	"errors"
	"testing"

	mockupdate "yadro.com/course/api/adapters/update/mock"
	updatepb "yadro.com/course/proto/update"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"

	"yadro.com/course/api/core"
)

func TestClient_Ping(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*mockupdate.MockUpdateClient)
		expectedErr string
	}{
		{
			name: "success",
			mockSetup: func(m *mockupdate.MockUpdateClient) {
				m.EXPECT().
					Ping(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
					Return(&emptypb.Empty{}, nil)
			},
		},
		{
			name: "error",
			mockSetup: func(m *mockupdate.MockUpdateClient) {
				m.EXPECT().
					Ping(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
					Return(nil, errors.New("network error"))
			},
			expectedErr: "failed to ping update service: network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mockupdate.NewMockUpdateClient(ctrl)
			tt.mockSetup(mockClient)

			client := &Client{
				client: mockClient,
			}

			err := client.Ping(context.Background())
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_Status(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*mockupdate.MockUpdateClient)
		expected    core.UpdateStatus
		expectedErr string
	}{
		{
			name: "idle status",
			mockSetup: func(m *mockupdate.MockUpdateClient) {
				m.EXPECT().
					Status(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
					Return(&updatepb.StatusReply{Status: updatepb.Status_STATUS_IDLE}, nil)
			},
			expected: core.StatusUpdateIdle,
		},
		{
			name: "running status",
			mockSetup: func(m *mockupdate.MockUpdateClient) {
				m.EXPECT().
					Status(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
					Return(&updatepb.StatusReply{Status: updatepb.Status_STATUS_RUNNING}, nil)
			},
			expected: core.StatusUpdateRunning,
		},
		{
			name: "unknown status",
			mockSetup: func(m *mockupdate.MockUpdateClient) {
				m.EXPECT().
					Status(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
					Return(&updatepb.StatusReply{Status: updatepb.Status(999)}, nil)
			},
			expected: core.StatusUpdateUnknown,
		},
		{
			name: "error",
			mockSetup: func(m *mockupdate.MockUpdateClient) {
				m.EXPECT().
					Status(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
					Return(nil, errors.New("status error"))
			},
			expectedErr: "failed to get status: status error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mockupdate.NewMockUpdateClient(ctrl)
			tt.mockSetup(mockClient)

			client := &Client{
				client: mockClient,
			}

			status, err := client.Status(context.Background())
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, status)
			}
		})
	}
}

func TestClient_Stats(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*mockupdate.MockUpdateClient)
		expected    core.UpdateStats
		expectedErr string
	}{
		{
			name: "success",
			mockSetup: func(m *mockupdate.MockUpdateClient) {
				m.EXPECT().
					Stats(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
					Return(&updatepb.StatsReply{
						WordsTotal:    100,
						WordsUnique:   80,
						ComicsFetched: 50,
						ComicsTotal:   200,
					}, nil)
			},
			expected: core.UpdateStats{
				WordsTotal:    100,
				WordsUnique:   80,
				ComicsFetched: 50,
				ComicsTotal:   200,
			},
		},
		{
			name: "error",
			mockSetup: func(m *mockupdate.MockUpdateClient) {
				m.EXPECT().
					Stats(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
					Return(nil, errors.New("stats error"))
			},
			expectedErr: "failed to get stats: stats error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mockupdate.NewMockUpdateClient(ctrl)
			tt.mockSetup(mockClient)

			client := &Client{
				client: mockClient,
			}

			stats, err := client.Stats(context.Background())
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, stats)
			}
		})
	}
}

func TestClient_Update(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*mockupdate.MockUpdateClient)
		expectedErr string
	}{
		{
			name: "success",
			mockSetup: func(m *mockupdate.MockUpdateClient) {
				m.EXPECT().
					Update(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
					Return(&emptypb.Empty{}, nil)
			},
		},
		{
			name: "error",
			mockSetup: func(m *mockupdate.MockUpdateClient) {
				m.EXPECT().
					Update(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
					Return(nil, errors.New("update error"))
			},
			expectedErr: "failed to update: update error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mockupdate.NewMockUpdateClient(ctrl)
			tt.mockSetup(mockClient)

			client := &Client{
				client: mockClient,
			}

			err := client.Update(context.Background())
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_Drop(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*mockupdate.MockUpdateClient)
		expectedErr string
	}{
		{
			name: "success",
			mockSetup: func(m *mockupdate.MockUpdateClient) {
				m.EXPECT().
					Drop(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
					Return(&emptypb.Empty{}, nil)
			},
		},
		{
			name: "error",
			mockSetup: func(m *mockupdate.MockUpdateClient) {
				m.EXPECT().
					Drop(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
					Return(nil, errors.New("drop error"))
			},
			expectedErr: "failed to drop database: drop error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mockupdate.NewMockUpdateClient(ctrl)
			tt.mockSetup(mockClient)

			client := &Client{
				client: mockClient,
			}

			err := client.Drop(context.Background())
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
