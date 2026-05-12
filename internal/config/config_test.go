package config

import "testing"

func TestParseReturnsDefaults(t *testing.T) {
	cfg, err := Parse(nil)
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}

	if cfg.ServerAddress != DefaultServerAddress {
		t.Fatalf("expected server address %q, got %q", DefaultServerAddress, cfg.ServerAddress)
	}
	if cfg.BaseURL != DefaultBaseURL {
		t.Fatalf("expected base URL %q, got %q", DefaultBaseURL, cfg.BaseURL)
	}
}

func TestParseUsesCommandLineFlags(t *testing.T) {
	cfg, err := Parse([]string{
		"-a", "localhost:8888",
		"-b", "http://localhost:8000",
	})
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}

	if cfg.ServerAddress != "localhost:8888" {
		t.Fatalf("expected server address %q, got %q", "localhost:8888", cfg.ServerAddress)
	}
	if cfg.BaseURL != "http://localhost:8000" {
		t.Fatalf("expected base URL %q, got %q", "http://localhost:8000", cfg.BaseURL)
	}
}
