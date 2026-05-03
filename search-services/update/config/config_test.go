package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"yadro.com/course/update/config"
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
update_address: localhost:8080
db_address: localhost:5432
words_address: localhost:81
broker_address: localhost:4222
xkcd:
  url: xkcd.com
  concurrency: 5
  timeout: 20s
  check_period: 2h
`)
		assert.NoError(t, err)
		assert.NoError(t, f.Close())

		cfg := config.MustLoad(f.Name())

		assert.Equal(t, "INFO", cfg.LogLevel)
		assert.Equal(t, "localhost:8080", cfg.Address)
		assert.Equal(t, "localhost:5432", cfg.DBAddress)
		assert.Equal(t, "xkcd.com", cfg.XKCD.URL)
		assert.Equal(t, 5, cfg.XKCD.Concurrency)
		assert.Equal(t, 20*time.Second, cfg.XKCD.Timeout)
		assert.Equal(t, 2*time.Hour, cfg.XKCD.CheckPeriod)
	})

	t.Run("fallback to env when file not found", func(t *testing.T) {
		t.Setenv("LOG_LEVEL", "WARN")
		t.Setenv("UPDATE_ADDRESS", "localhost:9090")
		t.Setenv("DB_ADDRESS", "localhost:5433")
		t.Setenv("WORDS_ADDRESS", "localhost:82")
		t.Setenv("BROKER_ADDRESS", "localhost:4223")
		t.Setenv("XKCD_URL", "xkcd.com")
		t.Setenv("XKCD_CONCURRENCY", "3")
		t.Setenv("XKCD_TIMEOUT", "30s")
		t.Setenv("XKCD_CHECK_PERIOD", "30m")

		cfg := config.MustLoad("nonexistent.yaml")

		assert.Equal(t, "WARN", cfg.LogLevel)
		assert.Equal(t, "localhost:9090", cfg.Address)
		assert.Equal(t, 3, cfg.XKCD.Concurrency)
		assert.Equal(t, 30*time.Second, cfg.XKCD.Timeout)
		assert.Equal(t, 30*time.Minute, cfg.XKCD.CheckPeriod)
	})
}