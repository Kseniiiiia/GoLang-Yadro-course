package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mockrest "yadro.com/course/api/adapters/rest/mock"
	"yadro.com/course/api/core"
)

func TestNewLoginHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mockrest.NewMockAuthenticator(ctrl)
	log := slog.Default()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful login",
			requestBody: map[string]string{
				"name":     "user",
				"password": "pass",
			},
			mockSetup: func() {
				mockAuth.EXPECT().
					Login("user", "pass").
					Return("valid_token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "valid_token",
		},
		{
			name: "invalid credentials",
			requestBody: map[string]string{
				"name":     "user",
				"password": "wrong",
			},
			mockSetup: func() {
				mockAuth.EXPECT().
					Login("user", "wrong").
					Return("", errors.New("invalid credentials"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request body\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler := NewLoginHandler(log, mockAuth)
			handler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}

func TestNewPingHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.Default()

	tests := []struct {
		name           string
		pingers        map[string]core.Pinger
		expectedStatus int
		expectedBody   PingResponse
	}{
		{
			name: "all services available",
			pingers: map[string]core.Pinger{
				"service1": createMockPinger(ctrl, nil),
				"service2": createMockPinger(ctrl, nil),
			},
			expectedStatus: http.StatusOK,
			expectedBody: PingResponse{
				Replies: map[string]string{
					"service1": "ok",
					"service2": "ok",
				},
			},
		},
		{
			name: "some services unavailable",
			pingers: map[string]core.Pinger{
				"ok":     createMockPinger(ctrl, nil),
				"failed": createMockPinger(ctrl, errors.New("unavailable")),
			},
			expectedStatus: http.StatusOK,
			expectedBody: PingResponse{
				Replies: map[string]string{
					"ok":     "ok",
					"failed": "unavailable",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/ping", nil)
			w := httptest.NewRecorder()

			handler := NewPingHandler(log, tt.pingers)
			handler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response PingResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)
		})
	}
}

func createMockPinger(ctrl *gomock.Controller, err error) *mockrest.MockPinger {
	mock := mockrest.NewMockPinger(ctrl)
	mock.EXPECT().Ping(gomock.Any()).Return(err).AnyTimes()
	return mock
}

func TestNewWordsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNorm := mockrest.NewMockNormalizer(ctrl)
	log := slog.Default()

	tests := []struct {
		name           string
		queryParams    map[string]string
		mockSetup      func()
		expectedStatus int
		expectedBody   WordsResponse
	}{
		{
			name: "successful normalization",
			queryParams: map[string]string{
				"phrase": "test phrase",
			},
			mockSetup: func() {
				mockNorm.EXPECT().
					Norm(gomock.Any(), "test phrase").
					Return([]string{"test", "phrase"}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: WordsResponse{
				Words: []string{"test", "phrase"},
				Total: 2,
			},
		},
		{
			name:           "missing phrase",
			queryParams:    map[string]string{},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "normalizer error",
			queryParams: map[string]string{
				"phrase": "test",
			},
			mockSetup: func() {
				mockNorm.EXPECT().
					Norm(gomock.Any(), "test").
					Return(nil, errors.New("normalization error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "bad arguments error",
			queryParams: map[string]string{
				"phrase": "test",
			},
			mockSetup: func() {
				mockNorm.EXPECT().
					Norm(gomock.Any(), "test").
					Return(nil, core.ErrBadArguments)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("GET", "/words", nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()

			handler := NewWordsHandler(log, mockNorm)
			handler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response WordsResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestNewUpdateStatsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := mockrest.NewMockUpdater(ctrl)
	log := slog.Default()

	tests := []struct {
		name           string
		mockSetup      func()
		expectedStatus int
		expectedBody   UpdateStatsResponse
	}{
		{
			name: "successful stats",
			mockSetup: func() {
				mockUpdater.EXPECT().
					Stats(gomock.Any()).
					Return(core.UpdateStats{
						WordsTotal:    100,
						WordsUnique:   50,
						ComicsFetched: 10,
						ComicsTotal:   200,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: UpdateStatsResponse{
				WordsTotal:    100,
				WordsUnique:   50,
				ComicsFetched: 10,
				ComicsTotal:   200,
			},
		},
		{
			name: "stats error",
			mockSetup: func() {
				mockUpdater.EXPECT().
					Stats(gomock.Any()).
					Return(core.UpdateStats{}, errors.New("stats error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("GET", "/stats", nil)
			w := httptest.NewRecorder()

			handler := NewUpdateStatsHandler(log, mockUpdater)
			handler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response UpdateStatsResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestNewUpdateStatusHandler(t *testing.T) {
	log := slog.Default()

	tests := []struct {
		name          string
		updateRunning bool
		expectedBody  UpdateStatusResponse
	}{
		{
			name:          "idle status",
			updateRunning: false,
			expectedBody:  UpdateStatusResponse{Status: "idle"},
		},
		{
			name:          "running status",
			updateRunning: true,
			expectedBody:  UpdateStatusResponse{Status: "running"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Используем реальный Updater, так как статус управляется atomic.Bool
			ctrl := gomock.NewController(t)
			mockUpdater := mockrest.NewMockUpdater(ctrl)
			ctrl.Finish()

			if tt.updateRunning {
				updates.Store(true)
			} else {
				updates.Store(false)
			}

			req := httptest.NewRequest("GET", "/status", nil)
			w := httptest.NewRecorder()

			handler := NewUpdateStatusHandler(log, mockUpdater)
			handler(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response UpdateStatusResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)
		})
	}
}

func TestNewUpdateHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := mockrest.NewMockUpdater(ctrl)
	log := slog.Default()

	tests := []struct {
		name           string
		initialState   bool
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:         "successful update",
			initialState: false,
			mockSetup: func() {
				mockUpdater.EXPECT().
					Update(gomock.Any()).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "update already running",
			initialState:   true,
			mockSetup:      func() {},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:         "update error",
			initialState: false,
			mockSetup: func() {
				mockUpdater.EXPECT().
					Update(gomock.Any()).
					Return(errors.New("update error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:         "already exists error",
			initialState: false,
			mockSetup: func() {
				mockUpdater.EXPECT().
					Update(gomock.Any()).
					Return(core.ErrAlreadyExists)
			},
			expectedStatus: http.StatusAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updates.Store(tt.initialState)
			defer updates.Store(false)

			tt.mockSetup()

			req := httptest.NewRequest("POST", "/update", nil)
			w := httptest.NewRecorder()

			handler := NewUpdateHandler(log, mockUpdater)
			handler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestNewDropHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := mockrest.NewMockUpdater(ctrl)
	log := slog.Default()

	tests := []struct {
		name           string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "successful drop",
			mockSetup: func() {
				mockUpdater.EXPECT().
					Drop(gomock.Any()).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "drop error",
			mockSetup: func() {
				mockUpdater.EXPECT().
					Drop(gomock.Any()).
					Return(errors.New("drop error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("POST", "/drop", nil)
			w := httptest.NewRecorder()

			handler := NewDropHandler(log, mockUpdater)
			handler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestNewSearchHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSearcher := mockrest.NewMockSearcher(ctrl)
	log := slog.Default()

	tests := []struct {
		name           string
		queryParams    map[string]string
		mockSetup      func()
		expectedStatus int
		expectedBody   SearchResponse
	}{
		{
			name: "successful search",
			queryParams: map[string]string{
				"phrase": "test",
				"limit":  "5",
			},
			mockSetup: func() {
				mockSearcher.EXPECT().
					Search(gomock.Any(), "test", int32(5)).
					Return([]core.Comics{{ID: 1, URL: "Test Comic"}}, int32(1), nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: SearchResponse{
				Comics: []core.Comics{{ID: 1, URL: "Test Comic"}},
				Total:  1,
			},
		},
		{
			name: "missing phrase",
			queryParams: map[string]string{
				"limit": "5",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid limit",
			queryParams: map[string]string{
				"phrase": "test",
				"limit":  "invalid",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "search error",
			queryParams: map[string]string{
				"phrase": "test",
				"limit":  "5",
			},
			mockSetup: func() {
				mockSearcher.EXPECT().
					Search(gomock.Any(), "test", int32(5)).
					Return(nil, int32(0), errors.New("search error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "bad arguments error",
			queryParams: map[string]string{
				"phrase": "test",
				"limit":  "5",
			},
			mockSetup: func() {
				mockSearcher.EXPECT().
					Search(gomock.Any(), "test", int32(5)).
					Return(nil, int32(0), core.ErrBadArguments)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("GET", "/search", nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()

			handler := NewSearchHandler(log, mockSearcher)
			handler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response SearchResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestNewSearchIndexHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSearcher := mockrest.NewMockSearcher(ctrl)
	log := slog.Default()

	tests := []struct {
		name           string
		queryParams    map[string]string
		mockSetup      func()
		expectedStatus int
		expectedBody   IndexSearchResponse
	}{
		{
			name: "successful index search",
			queryParams: map[string]string{
				"phrase": "test",
				"limit":  "5",
			},
			mockSetup: func() {
				mockSearcher.EXPECT().
					IndexSearch(gomock.Any(), "test", int32(5)).
					Return([]core.Comics{{ID: 1, URL: "Test Comic"}}, int32(1), nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: IndexSearchResponse{
				Comics: []core.Comics{{ID: 1, URL: "Test Comic"}},
				Total:  1,
			},
		},
		{
			name: "missing phrase",
			queryParams: map[string]string{
				"limit": "5",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid limit",
			queryParams: map[string]string{
				"phrase": "test",
				"limit":  "invalid",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "index search error",
			queryParams: map[string]string{
				"phrase": "test",
				"limit":  "5",
			},
			mockSetup: func() {
				mockSearcher.EXPECT().
					IndexSearch(gomock.Any(), "test", int32(5)).
					Return(nil, int32(0), errors.New("search error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "bad arguments error",
			queryParams: map[string]string{
				"phrase": "test",
				"limit":  "5",
			},
			mockSetup: func() {
				mockSearcher.EXPECT().
					IndexSearch(gomock.Any(), "test", int32(5)).
					Return(nil, int32(0), core.ErrBadArguments)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("GET", "/index-search", nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()

			handler := NewSearchIndexHandler(log, mockSearcher)
			handler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response IndexSearchResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}
