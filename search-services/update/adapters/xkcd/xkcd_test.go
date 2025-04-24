package xkcd

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	mockxkcd "yadro.com/course/update/adapters/xkcd/mock"
	"yadro.com/course/update/core"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		timeout     time.Duration
		expectedErr string
	}{
		{
			name:    "successful creation",
			url:     "http://example.com",
			timeout: time.Second,
		},
		{
			name:        "empty url",
			url:         "",
			timeout:     time.Second,
			expectedErr: "empty base url specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.url, tt.timeout, nil)
			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestClient_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/123/info.0.json":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"num":        123,
				"img":        "http://example.com/123.png",
				"title":      "Test Comic",
				"transcript": "T",
				"alt":        " ",
			})
		case "/404/info.0.json":
			w.WriteHeader(http.StatusNotFound)
		case "/500/info.0.json":
			w.WriteHeader(http.StatusInternalServerError)
		case "/invalid/info.0.json":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("invalid json"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	tests := []struct {
		name        string
		id          int
		expected    core.XKCDInfo
		expectedErr string
	}{
		{
			name: "successful get",
			id:   123,
			expected: core.XKCDInfo{
				NUM:         123,
				URL:         "http://example.com/123.png",
				Title:       "Test Comic",
				Description: "T Test Comic",
			},
		},
		{
			name:        "not found",
			id:          404,
			expectedErr: core.ErrNotFound.Error(),
		},
		{
			name:        "server error",
			id:          500,
			expectedErr: "status 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(ts.URL, time.Second, nil)
			assert.NoError(t, err)

			var id int
			if tt.id == -1 {
				id = 999
			} else {
				id = tt.id
			}

			result, err := client.Get(context.Background(), id)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				if errors.Is(err, core.ErrNotFound) {
					missing := client.MissingIds(context.Background())
					assert.Contains(t, missing, id)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestClient_LastID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/info.0.json":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"num":        1000,
				"img":        "http://example.com/1000.png",
				"title":      "Last Comic",
				"transcript": "Last transcript",
			})
		case "/error/info.0.json":
			w.WriteHeader(http.StatusInternalServerError)
		case "/invalid/info.0.json":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("invalid json"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	tests := []struct {
		name        string
		baseURL     string
		expected    int
		expectedErr string
	}{
		{
			name:     "successful last id",
			baseURL:  ts.URL,
			expected: 1000,
		},
		{
			name:        "server error",
			baseURL:     ts.URL + "/error",
			expectedErr: "status 500",
		},
		{
			name:        "invalid json",
			baseURL:     ts.URL + "/invalid",
			expectedErr: "failed to decode last comic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.baseURL, time.Second, nil)
			assert.NoError(t, err)

			result, err := client.LastID(context.Background())

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

func TestClient_MissingIds(t *testing.T) {
	t.Run("empty missing ids", func(t *testing.T) {
		client, err := NewClient("http://example.com", time.Second, nil)
		assert.NoError(t, err)

		missing := client.MissingIds(context.Background())
		assert.Empty(t, missing)
	})

	t.Run("with missing ids", func(t *testing.T) {
		client, err := NewClient("http://example.com", time.Second, nil)
		assert.NoError(t, err)

		_, _ = client.Get(context.Background(), 404)

		missing := client.MissingIds(context.Background())
		assert.Contains(t, missing, 404)
		assert.Len(t, missing, 1)
	})
}

func TestXKCDInterface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockXKCD := mockxkcd.NewMockXKCD(ctrl)

	t.Run("Get with mock", func(t *testing.T) {
		expected := core.XKCDInfo{
			NUM:         123,
			URL:         "http://example.com/123.png",
			Title:       "Test Comic",
			Description: "Test Description",
		}

		mockXKCD.EXPECT().
			Get(gomock.Any(), 123).
			Return(expected, nil)

		result, err := mockXKCD.Get(context.Background(), 123)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("LastID with mock", func(t *testing.T) {
		mockXKCD.EXPECT().
			LastID(gomock.Any()).
			Return(1000, nil)

		result, err := mockXKCD.LastID(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 1000, result)
	})

	t.Run("MissingIds with mock", func(t *testing.T) {
		expected := []int{404, 405}
		mockXKCD.EXPECT().
			MissingIds(gomock.Any()).
			Return(expected)

		result := mockXKCD.MissingIds(context.Background())
		assert.Equal(t, expected, result)
	})
}
