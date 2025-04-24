package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type SEARCHConfig struct {
	Address  string        `yaml:"address" env:"SEARCH_ADDRESS" env-default:"localhost:80"`
	Timeout  time.Duration `yaml:"timeout" env:"SEARCH_TIMEOUT" env-default:"5s"`
	IndexTTL time.Duration `yaml:"index_ttl" env:"INDEX_TTL" env-default:"20s"`
}

type Config struct {
	LogLevel     string       `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	SearchConfig SEARCHConfig `yaml:"search_server"`
	DBAddress    string       `yaml:"db_address" env:"DB_ADDRESS" env-default:"postgres://user:password@localhost:5432/dbname"`
	WordsAddress string       `yaml:"words_address" env:"WORDS_ADDRESS" env-default:"localhost:50051"`
}

func MustLoad(configPath string) Config {
	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return cfg
}
