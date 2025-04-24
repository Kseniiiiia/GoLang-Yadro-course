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
search_server:
  address: "search-service:8080"
  timeout: 10s
  index_ttl: 30s
db_address: "postgres://user:pass@db-service:5432/dbname"
words_address: "words-service:8081"
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
		assert.Equal(t, "search-service:8080", cfg.SearchConfig.Address)
		assert.Equal(t, 10*time.Second, cfg.SearchConfig.Timeout)
		assert.Equal(t, 30*time.Second, cfg.SearchConfig.IndexTTL)
		assert.Equal(t, "postgres://user:pass@db-service:5432/dbname", cfg.DBAddress)
		assert.Equal(t, "words-service:8081", cfg.WordsAddress)
	})

	t.Run("override with env vars", func(t *testing.T) {
		t.Setenv("LOG_LEVEL", "DEBUG")
		t.Setenv("SEARCH_ADDRESS", ":9090")
		t.Setenv("SEARCH_TIMEOUT", "15s")
		t.Setenv("DB_ADDRESS", "postgres://envuser:envpass@localhost:5432/envdb")
		defer func() {
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("SEARCH_ADDRESS")
			os.Unsetenv("SEARCH_TIMEOUT")
			os.Unsetenv("DB_ADDRESS")
		}()

		cfg := MustLoad(tmpFile.Name())

		assert.Equal(t, "DEBUG", cfg.LogLevel)
		assert.Equal(t, ":9090", cfg.SearchConfig.Address)
		assert.Equal(t, 15*time.Second, cfg.SearchConfig.Timeout)
		assert.Equal(t, "postgres://envuser:envpass@localhost:5432/envdb", cfg.DBAddress)

		// Проверяем значения, которые не переопределялись
		assert.Equal(t, 30*time.Second, cfg.SearchConfig.IndexTTL)
		assert.Equal(t, "words-service:8081", cfg.WordsAddress)
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
		assert.Equal(t, "localhost:80", cfg.SearchConfig.Address)
		assert.Equal(t, 5*time.Second, cfg.SearchConfig.Timeout)
		assert.Equal(t, 20*time.Second, cfg.SearchConfig.IndexTTL)
		assert.Equal(t, "postgres://user:password@localhost:5432/dbname", cfg.DBAddress)
		assert.Equal(t, "localhost:50051", cfg.WordsAddress)
	})
}

func TestSEARCHConfig(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		cfg := SEARCHConfig{}
		assert.Equal(t, "", cfg.Address)
		assert.Equal(t, time.Duration(0), cfg.Timeout)
		assert.Equal(t, time.Duration(0), cfg.IndexTTL)
	})
}
