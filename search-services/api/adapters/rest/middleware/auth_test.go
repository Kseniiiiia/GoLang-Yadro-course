package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var errTestInvalidToken = errors.New("invalid token")

type MockTokenVerifier struct {
	err error
}

func (m *MockTokenVerifier) Verify(token string) error {
	return m.err
}

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "missing auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Authorization header is missing",
		},
		{
			name:           "invalid auth format",
			authHeader:     "Bearer token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid authorization format. Expected 'Token <jwt>'",
		},
		{
			name:           "empty token",
			authHeader:     "Token ",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Empty token",
		},
		{
			name:           "invalid token",
			authHeader:     "Token bad_token",
			mockError:      errTestInvalidToken,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid token: " + errTestInvalidToken.Error(),
		},
		{
			name:           "valid token",
			authHeader:     "Token good_token",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier := &MockTokenVerifier{err: tt.mockError}
			handler := Auth(
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("OK"))
				},
				verifier,
			)

			req := httptest.NewRequest("GET", "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if tt.expectedBody != "" {
				body := strings.TrimSpace(rr.Body.String())
				if body != tt.expectedBody {
					t.Errorf("handler returned unexpected body: got %v want %v",
						body, tt.expectedBody)
				}
			}
		})
	}
}
