package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"yadro.com/course/api/config"
)

func TestMustLoad(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		f, err := os.CreateTemp("", "config-*.yaml")
		assert.NoError(t, err)
		t.Cleanup(func() {
			assert.NoError(t, os.Remove(f.Name()))
		})

		_, err = f.WriteString(`
log_level: INFO
search_concurrency: 5
search_rate: 10
token_ttl: 1h
api_server:
  address: localhost:8080
  timeout: 10s
words_address: words:81
update_address: update:82
search_address: search:83
`)
		assert.NoError(t, err)
		assert.NoError(t, f.Close())

		cfg := config.MustLoad(f.Name())

		assert.Equal(t, "INFO", cfg.LogLevel)
		assert.Equal(t, 5, cfg.SearchConcurrency)
		assert.Equal(t, 10, cfg.SearchRate)
		assert.Equal(t, time.Hour, cfg.TokenTTL)
		assert.Equal(t, "localhost:8080", cfg.HTTPConfig.Address)
		assert.Equal(t, 10*time.Second, cfg.HTTPConfig.Timeout)
	})
}