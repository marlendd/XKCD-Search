package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"yadro.com/course/search/config"
)

func TestMustLoad(t *testing.T) {
	t.Run("valid config file", func(t *testing.T) {
		f, err := os.CreateTemp("", "config-*.yaml")
		assert.NoError(t, err)
		t.Cleanup(func() {
			assert.NoError(t, os.Remove(f.Name()))
		})

		_, err = f.WriteString(`
log_level: INFO
search_address: localhost:8080
db_address: localhost:5432
words_address: localhost:81
index_ttl: 30s
broker_address: localhost:4222
`)
		assert.NoError(t, err)
		assert.NoError(t, f.Close())

		cfg := config.MustLoad(f.Name())

		assert.Equal(t, "INFO", cfg.LogLevel)
		assert.Equal(t, "localhost:8080", cfg.Address)
		assert.Equal(t, "localhost:5432", cfg.DBAddress)
		assert.Equal(t, 30*time.Second, cfg.IndexTTL)
	})

	t.Run("fallback to env when file not found", func(t *testing.T) {
		t.Setenv("LOG_LEVEL", "WARN")
		t.Setenv("SEARCH_ADDRESS", "localhost:9090")
		t.Setenv("DB_ADDRESS", "localhost:5433")
		t.Setenv("WORDS_ADDRESS", "localhost:82")
		t.Setenv("INDEX_TTL", "1m")
		t.Setenv("BROKER_ADDRESS", "localhost:4223")

		cfg := config.MustLoad("nonexistent.yaml")

		assert.Equal(t, "WARN", cfg.LogLevel)
		assert.Equal(t, "localhost:9090", cfg.Address)
		assert.Equal(t, "localhost:5433", cfg.DBAddress)
		assert.Equal(t, time.Minute, cfg.IndexTTL)
	})
}