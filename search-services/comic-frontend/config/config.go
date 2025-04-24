package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Server struct {
	Address string        `yaml:"address" env:"SERVER_ADDRESS" env-default:"http_server:84"`
	Timeout time.Duration `yaml:"timeout" env:"HTTP_TIMEOUT" env-default:"5s"`
}

type API struct {
	ApiURL string `yaml:"url" env:"API_BASE_URL" env-default:"localhost:28080"`
}

type Config struct {
	LogLevel string `yaml:"log_level" env:"LOG_LEVEL" env-default:"INFO"`
	Server   Server `yaml:"server"`
	API      API    `yaml:"api"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
