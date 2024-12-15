package server

import (
	"flag"
	"strings"

	"github.com/caarlos0/env/v11"
)

const (
	defaultAddress     = "localhost:8080"
	defaultLogLevel    = "info"
	defaultPostgresDSN = "postgresql://postgres:password@localhost:5432/keeper"

	logUsage    = "log level (default 'info')"
	addresUsage = "server address (default localhost:8080)"
	dsnUsage    = "postgres connection string (default postgresql://postgres:password@localhost:5432/keeper)"
)

type Config struct {
	Address     string `env:"ADDRESS"`
	LogLevel    string `env:"LOG_LEVEL"`
	PostgresDSN string `env:"POSTGRES_DSN"`
}

func NewConfig() (*Config, error) {
	cfg := Config{
		Address:     defaultAddress,
		LogLevel:    defaultLogLevel,
		PostgresDSN: defaultPostgresDSN,
	}

	flag.StringVar(&cfg.Address, "a", defaultAddress, addresUsage)
	flag.StringVar(&cfg.LogLevel, "l", defaultLogLevel, logUsage)
	flag.StringVar(&cfg.PostgresDSN, "d", defaultPostgresDSN, dsnUsage)

	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	if strings.HasPrefix(cfg.Address, "http://") {
		cfg.Address = strings.Replace(cfg.Address, "http://", "", -1)
	}

	return &cfg, nil
}
