package configs

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTP   Http
	Logger Logger
}

type Http struct {
	StartPage  string `yaml:"start_page" env:"START_PAGE" env-default:"http://193.168.227.93/start"`
	MaxRPC     int    `yaml:"max_rpc" env:"MAX_RPC" env-default:"3"`
	ThreadsNum int    `yaml:"threads_num" env:"THREADS_NUM" env-default:"5"`
}

type Logger struct {
	Level string `yaml:"level" env:"LOGGER_LEVEL" env-default:"info"`
}

func New() (*Config, error) {
	cfg := &Config{}

	if err := cleanenv.ReadConfig("configs/configs.yaml", cfg); err != nil {
		return nil, fmt.Errorf(".yaml: %w", err)
	}

	return cfg, nil
}
