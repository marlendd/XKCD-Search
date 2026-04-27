package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type XKCD struct {
	URL         string        `yaml:"url" env:"XKCD_URL" env-default:"xkcd.com"`
	Concurrency int           `yaml:"concurrency" env:"XKCD_CONCURRENCY" env-default:"1"`
	Timeout     time.Duration `yaml:"timeout" env:"XKCD_TIMEOUT" env-default:"10s"`
	CheckPeriod time.Duration `yaml:"check_period" env:"XKCD_CHECK_PERIOD" env-default:"1h"`
}

type Config struct {
	LogLevel      string `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	Address       string `yaml:"update_address" env:"UPDATE_ADDRESS" env-default:"localhost:80"`
	XKCD          XKCD   `yaml:"xkcd"`
	DBAddress     string `yaml:"db_address" env:"DB_ADDRESS" env-default:"localhost:82"`
	WordsAddress  string `yaml:"words_address" env:"WORDS_ADDRESS" env-default:"localhost:81"`
	BrokerAddress string `yaml:"broker_address" env:"BROKER_ADDRESS" env-default:"localhost:4222"`
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
