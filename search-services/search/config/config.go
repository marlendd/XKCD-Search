package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel      string        `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	Address       string        `yaml:"search_address" env:"SEARCH_ADDRESS" env-default:"localhost:80"`
	DBAddress     string        `yaml:"db_address" env:"DB_ADDRESS" env-default:"localhost:82"`
	WordsAddress  string        `yaml:"words_address" env:"WORDS_ADDRESS" env-default:"localhost:81"`
	IndexTTL      time.Duration `yaml:"index_ttl" env:"INDEX_TTL" env-default:"20s"`
	BrokerAddress string        `yaml:"broker_address" env:"BROKER_ADDRESS" env-default:"localhost:4222"`
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
