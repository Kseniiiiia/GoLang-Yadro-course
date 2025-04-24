package aaa

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		os.Setenv("ADMIN_USER", "admin")
		os.Setenv("ADMIN_PASSWORD", "password")
		defer func() {
			os.Unsetenv("ADMIN_USER")
			os.Unsetenv("ADMIN_PASSWORD")
		}()

		service, err := New(1*time.Hour, slog.Default())
		assert.NoError(t, err)
		assert.NotNil(t, service)
	})

	t.Run("missing admin user", func(t *testing.T) {
		os.Unsetenv("ADMIN_USER")
		os.Setenv("ADMIN_PASSWORD", "password")
		defer os.Unsetenv("ADMIN_PASSWORD")

		_, err := New(1*time.Hour, slog.Default())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "admin user")
	})

	t.Run("missing admin password", func(t *testing.T) {
		os.Setenv("ADMIN_USER", "admin")
		os.Unsetenv("ADMIN_PASSWORD")
		defer os.Unsetenv("ADMIN_USER")

		_, err := New(1*time.Hour, slog.Default())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "admin password")
	})
}

func TestLogin(t *testing.T) {
	os.Setenv("ADMIN_USER", "admin")
	os.Setenv("ADMIN_PASSWORD", "password")
	defer func() {
		os.Unsetenv("ADMIN_USER")
		os.Unsetenv("ADMIN_PASSWORD")
	}()

	service, _ := New(1*time.Hour, slog.Default())

	t.Run("successful login", func(t *testing.T) {
		token, err := service.Login("admin", "password")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		_, err = jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
		assert.NoError(t, err)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		_, err := service.Login("admin", "wrong")
		assert.Error(t, err)
		assert.Equal(t, "invalid credentials", err.Error())

		_, err = service.Login("wrong", "password")
		assert.Error(t, err)
		assert.Equal(t, "invalid credentials", err.Error())
	})
}

func TestVerify(t *testing.T) {
	os.Setenv("ADMIN_USER", "admin")
	os.Setenv("ADMIN_PASSWORD", "password")
	defer func() {
		os.Unsetenv("ADMIN_USER")
		os.Unsetenv("ADMIN_PASSWORD")
	}()

	service, _ := New(1*time.Hour, slog.Default())
	token, _ := service.Login("admin", "password")

	t.Run("valid token", func(t *testing.T) {
		err := service.Verify(token)
		assert.NoError(t, err)
	})

	t.Run("invalid token", func(t *testing.T) {
		err := service.Verify("invalid token")
		assert.Error(t, err)
		assert.Equal(t, "invalid token", err.Error())
	})

	t.Run("expired token", func(t *testing.T) {
		expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": adminRole,
			"exp": time.Now().Add(-1 * time.Hour).Unix(),
		})
		tokenString, _ := expiredToken.SignedString([]byte(secretKey))

		err := service.Verify(tokenString)
		assert.Error(t, err)
		assert.Equal(t, "invalid token", err.Error())
	})

	t.Run("invalid subject", func(t *testing.T) {
		wrongSubToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "wrong_role",
			"exp": time.Now().Add(1 * time.Hour).Unix(),
		})
		tokenString, _ := wrongSubToken.SignedString([]byte(secretKey))

		err := service.Verify(tokenString)
		assert.Error(t, err)
		assert.Equal(t, "not authorized", err.Error())
	})
}
