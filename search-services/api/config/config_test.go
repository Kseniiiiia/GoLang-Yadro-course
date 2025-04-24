package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMustLoad(t *testing.T) {

	configContent := `
log_level: INFO
search_concurrency: 2
search_rate: 3
api_server:
  address: ":8080"
  timeout: 10s
words_address: "words-service:81"
update_address: "update-service:82"
search_address: "search-service:83"
token_ttl: 12h
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	t.Run("load from file", func(t *testing.T) {
		cfg := MustLoad(tmpFile.Name())

		assert.Equal(t, "INFO", cfg.LogLevel)
		assert.Equal(t, 2, cfg.SearchConcurrency)
		assert.Equal(t, 3, cfg.SearchRate)
		assert.Equal(t, ":8080", cfg.HTTPConfig.Address)
		assert.Equal(t, 10*time.Second, cfg.HTTPConfig.Timeout)
		assert.Equal(t, "words-service:81", cfg.WordsAddress)
		assert.Equal(t, "update-service:82", cfg.UpdateAddress)
		assert.Equal(t, "search-service:83", cfg.SearchAddress)
		assert.Equal(t, 12*time.Hour, cfg.TokenTTL)
	})

	t.Run("override with env vars", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "DEBUG")
		os.Setenv("API_ADDRESS", ":9090")
		os.Setenv("TOKEN_TTL", "1h")
		defer func() {
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("API_ADDRESS")
			os.Unsetenv("TOKEN_TTL")
		}()

		cfg := MustLoad(tmpFile.Name())

		assert.Equal(t, "DEBUG", cfg.LogLevel)
		assert.Equal(t, ":9090", cfg.HTTPConfig.Address)
		assert.Equal(t, time.Hour, cfg.TokenTTL)

		assert.Equal(t, 2, cfg.SearchConcurrency)
		assert.Equal(t, 3, cfg.SearchRate)
		assert.Equal(t, 10*time.Second, cfg.HTTPConfig.Timeout)
		assert.Equal(t, "words-service:81", cfg.WordsAddress)
	})

	t.Run("empty config file should use defaults", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "empty-config-*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString("{}\n")
		require.NoError(t, err)
		require.NoError(t, tmpFile.Close())

		cfg := MustLoad(tmpFile.Name())

		assert.Equal(t, "DEBUG", cfg.LogLevel)
		assert.Equal(t, 1, cfg.SearchConcurrency)
		assert.Equal(t, 1, cfg.SearchRate)
		assert.Equal(t, "localhost:80", cfg.HTTPConfig.Address)
		assert.Equal(t, 5*time.Second, cfg.HTTPConfig.Timeout)
		assert.Equal(t, "words:81", cfg.WordsAddress)
		assert.Equal(t, "update:82", cfg.UpdateAddress)
		assert.Equal(t, "search:83", cfg.SearchAddress)
		assert.Equal(t, 24*time.Hour, cfg.TokenTTL)
	})
}

func TestHTTPConfig(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		cfg := HTTPConfig{}
		assert.Equal(t, "", cfg.Address)
		assert.Equal(t, time.Duration(0), cfg.Timeout)
	})
}
