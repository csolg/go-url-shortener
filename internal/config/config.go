package config

import (
	"flag"
	"io"
	"os"
)

const (
	DefaultServerAddress = "localhost:8080"
	DefaultBaseURL       = "http://localhost:8080"
	DefaultDatabaseDSN   = "shortener.db"
)

type Config struct {
	ServerAddress string
	BaseURL       string
	DatabaseDSN   string
}

func Parse(args []string) (Config, error) {
	cfg := Config{
		ServerAddress: DefaultServerAddress,
		BaseURL:       DefaultBaseURL,
		DatabaseDSN:   databaseDSN(),
	}

	flags := flag.NewFlagSet("shortener", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server address")
	flags.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "base URL for shortened links")

	if err := flags.Parse(args); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func databaseDSN() string {
	if dsn := os.Getenv("DATABASE_DSN"); dsn != "" {
		return dsn
	}

	return DefaultDatabaseDSN
}
