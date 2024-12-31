package client

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

const (
	defaultAddress     = "http://localhost:8080"
	defaultPostgresDSN = "postgresql://postgres:password@localhost:5432/keeper"
	defaultKeyPath     = "keys/"

	addressUsage = "server address (default localhost:8080)"
	dsnUsage     = "postgres connection string (default postgresql://postgres:password@localhost:5432/keeper)"
	keysUsage    = "path to user keys (default keys/)"
)

type Config struct {
	Address     string `env:"ADDRESS"`
	PostgresDSN string `env:"POSTGRES_DSN"`
	KeyPath     string `env:"KEY_PATH"`
}

func NewConfig() (*Config, error) {
	cfg := Config{
		Address:     defaultAddress,
		PostgresDSN: defaultPostgresDSN,
		KeyPath:     defaultKeyPath,
	}

	flag.StringVar(&cfg.Address, "a", defaultAddress, addressUsage)
	flag.StringVar(&cfg.PostgresDSN, "d", defaultPostgresDSN, dsnUsage)
	flag.StringVar(&cfg.KeyPath, "k", defaultKeyPath, keysUsage)

	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
