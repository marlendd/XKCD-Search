package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel   string        `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	Address    string        `yaml:"frontend_address" env:"FRONTEND_ADDRESS" env-default:"localhost:8080"`
	TokenTTL   time.Duration `yaml:"token_ttl" env:"TOKEN_TTL" env-default:"24h"`
	BaseURL    string        `yaml:"base_url" env:"BASE_URL" env-default:"http://localhost:28080"`
	Timeout    time.Duration `yaml:"timeout" env:"TIMEOUT" env-default:"5s"`
}

func MustLoad(configPath string) Config {
	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		if os.IsNotExist(err) {
			if err := cleanenv.ReadEnv(&cfg); err != nil {
				slog.Error("failed to read env", "error", err)
				os.Exit(1)
			}
		} else {
			slog.Error("failed to read config file", "error", err)
			os.Exit(1)
		}
	}

	return cfg
}
