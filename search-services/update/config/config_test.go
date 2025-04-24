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
update_address: "update-service:8080"
db_address: "db-service:5432"
words_address: "words-service:8081"
xkcd:
  url: "https://xkcd-api.com"
  concurrency: 5
  timeout: 30s
  check_period: 2h
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
		assert.Equal(t, "update-service:8080", cfg.Address)
		assert.Equal(t, "db-service:5432", cfg.DBAddress)
		assert.Equal(t, "words-service:8081", cfg.WordsAddress)

		// Проверяем XKCD конфигурацию
		assert.Equal(t, "https://xkcd-api.com", cfg.XKCD.URL)
		assert.Equal(t, 5, cfg.XKCD.Concurrency)
		assert.Equal(t, 30*time.Second, cfg.XKCD.Timeout)
		assert.Equal(t, 2*time.Hour, cfg.XKCD.CheckPeriod)
	})

	t.Run("override with env vars", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "DEBUG")
		os.Setenv("UPDATE_ADDRESS", ":9090")
		os.Setenv("XKCD_URL", "https://alt-xkcd-api.com")
		os.Setenv("XKCD_TIMEOUT", "15s")
		defer func() {
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("UPDATE_ADDRESS")
			os.Unsetenv("XKCD_URL")
			os.Unsetenv("XKCD_TIMEOUT")
		}()

		cfg := MustLoad(tmpFile.Name())

		assert.Equal(t, "DEBUG", cfg.LogLevel)
		assert.Equal(t, ":9090", cfg.Address)
		assert.Equal(t, "https://alt-xkcd-api.com", cfg.XKCD.URL)
		assert.Equal(t, 15*time.Second, cfg.XKCD.Timeout)

		assert.Equal(t, "db-service:5432", cfg.DBAddress)
		assert.Equal(t, "words-service:8081", cfg.WordsAddress)
		assert.Equal(t, 5, cfg.XKCD.Concurrency)
		assert.Equal(t, 2*time.Hour, cfg.XKCD.CheckPeriod)
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
		assert.Equal(t, "localhost:80", cfg.Address)
		assert.Equal(t, "localhost:82", cfg.DBAddress)
		assert.Equal(t, "localhost:81", cfg.WordsAddress)

		assert.Equal(t, "xkcd.com", cfg.XKCD.URL)
		assert.Equal(t, 1, cfg.XKCD.Concurrency)
		assert.Equal(t, 10*time.Second, cfg.XKCD.Timeout)
		assert.Equal(t, 1*time.Hour, cfg.XKCD.CheckPeriod)
	})
}

func TestXKCDConfig(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		cfg := XKCD{}
		assert.Equal(t, "", cfg.URL)
		assert.Equal(t, 0, cfg.Concurrency)
		assert.Equal(t, time.Duration(0), cfg.Timeout)
		assert.Equal(t, time.Duration(0), cfg.CheckPeriod)
	})
}
