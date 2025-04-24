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

	updatepb "yadro.com/course/proto/update"
	mockserver "yadro.com/course/update/adapters/grpc/mock"
	"yadro.com/course/update/core"
)

func TestNewServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mockserver.NewMockUpdater(ctrl)
	server := NewServer(mockService)

	assert.NotNil(t, server)
	assert.Equal(t, mockService, server.service)
}

func TestServer_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mockserver.NewMockUpdater(ctrl)
	server := NewServer(mockService)

	resp, err := server.Ping(context.Background(), &emptypb.Empty{})

	assert.NoError(t, err)
	assert.IsType(t, &emptypb.Empty{}, resp)
}

func TestServer_Status(t *testing.T) {
	tests := []struct {
		name           string
		serviceStatus  core.ServiceStatus
		expectedStatus updatepb.Status
	}{
		{
			name:           "Idle status",
			serviceStatus:  "idle",
			expectedStatus: updatepb.Status_STATUS_IDLE,
		},
		{
			name:           "Running status",
			serviceStatus:  "running",
			expectedStatus: updatepb.Status_STATUS_RUNNING,
		},
		{
			name:           "Unknown status",
			serviceStatus:  "unknown",
			expectedStatus: updatepb.Status_STATUS_UNSPECIFIED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mockserver.NewMockUpdater(ctrl)
			mockService.EXPECT().Status(gomock.Any()).Return(tt.serviceStatus)

			server := NewServer(mockService)
			resp, err := server.Status(context.Background(), &emptypb.Empty{})

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.Status)
		})
	}
}

func TestServer_Update(t *testing.T) {
	tests := []struct {
		name         string
		mockSetup    func(*mockserver.MockUpdater)
		expectedErr  error
		expectedCode codes.Code
	}{
		{
			name: "Successful update",
			mockSetup: func(m *mockserver.MockUpdater) {
				m.EXPECT().Update(gomock.Any()).Return(nil)
			},
			expectedErr:  nil,
			expectedCode: codes.OK,
		},
		{
			name: "Already exists error",
			mockSetup: func(m *mockserver.MockUpdater) {
				m.EXPECT().Update(gomock.Any()).Return(core.ErrAlreadyExists)
			},
			expectedErr:  core.ErrAlreadyExists,
			expectedCode: codes.AlreadyExists,
		},
		{
			name: "Internal error",
			mockSetup: func(m *mockserver.MockUpdater) {
				m.EXPECT().Update(gomock.Any()).Return(errors.New("some error"))
			},
			expectedErr:  errors.New("some error"),
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mockserver.NewMockUpdater(ctrl)
			tt.mockSetup(mockService)

			server := NewServer(mockService)
			resp, err := server.Update(context.Background(), &emptypb.Empty{})

			if tt.expectedErr != nil {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Contains(t, st.Message(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.IsType(t, &emptypb.Empty{}, resp)
			}
		})
	}
}

func TestServer_Stats(t *testing.T) {
	tests := []struct {
		name         string
		mockSetup    func(*mockserver.MockUpdater)
		expectedResp *updatepb.StatsReply
		expectedErr  error
	}{
		{
			name: "Successful stats retrieval",
			mockSetup: func(m *mockserver.MockUpdater) {
				m.EXPECT().Stats(gomock.Any()).Return(core.ServiceStats{
					DBStats: core.DBStats{
						WordsTotal:    100,
						WordsUnique:   80,
						ComicsFetched: 40,
					},
					ComicsTotal: 50,
				}, nil)
			},
			expectedResp: &updatepb.StatsReply{
				WordsTotal:    100,
				WordsUnique:   80,
				ComicsTotal:   50,
				ComicsFetched: 40,
			},
			expectedErr: nil,
		},
		{
			name: "Error getting stats",
			mockSetup: func(m *mockserver.MockUpdater) {
				m.EXPECT().Stats(gomock.Any()).Return(core.ServiceStats{}, errors.New("stats error"))
			},
			expectedResp: nil,
			expectedErr:  errors.New("stats error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mockserver.NewMockUpdater(ctrl)
			tt.mockSetup(mockService)

			server := NewServer(mockService)
			resp, err := server.Stats(context.Background(), &emptypb.Empty{})

			if tt.expectedErr != nil {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, codes.Internal, st.Code())
				assert.Contains(t, st.Message(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}

func TestServer_Drop(t *testing.T) {
	tests := []struct {
		name         string
		mockSetup    func(*mockserver.MockUpdater)
		expectedErr  error
		expectedCode codes.Code
	}{
		{
			name: "Successful drop",
			mockSetup: func(m *mockserver.MockUpdater) {
				m.EXPECT().Drop(gomock.Any()).Return(nil)
			},
			expectedErr:  nil,
			expectedCode: codes.OK,
		},
		{
			name: "Error during drop",
			mockSetup: func(m *mockserver.MockUpdater) {
				m.EXPECT().Drop(gomock.Any()).Return(errors.New("drop error"))
			},
			expectedErr:  errors.New("drop error"),
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mockserver.NewMockUpdater(ctrl)
			tt.mockSetup(mockService)

			server := NewServer(mockService)
			resp, err := server.Drop(context.Background(), &emptypb.Empty{})

			if tt.expectedErr != nil {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Contains(t, st.Message(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.IsType(t, &emptypb.Empty{}, resp)
			}
		})
	}
}

func TestToProtoStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    core.ServiceStatus
		expected updatepb.Status
	}{
		{"Idle status", "idle", updatepb.Status_STATUS_IDLE},
		{"Running status", "running", updatepb.Status_STATUS_RUNNING},
		{"Unknown status", "unknown", updatepb.Status_STATUS_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toProtoStatus(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
