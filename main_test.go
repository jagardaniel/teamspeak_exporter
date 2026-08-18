package main

import (
	"testing"

	"github.com/alecthomas/kong"
)

func parseConfig(t *testing.T, args ...string) (*Config, error) {
	t.Helper()

	cfg := &Config{}

	parser, err := kong.New(cfg)
	if err != nil {
		t.Fatalf("failed to create kong parser: %v", err)
	}

	_, err = parser.Parse(args)

	return cfg, err
}

func TestDefaults(t *testing.T) {
	// API-key is required so just set it to prevent error
	cfg, err := parseConfig(t, "--api-key", "test-key")
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Listen != ":9289" {
		t.Errorf("expected ':9289', got %q", cfg.Listen)
	}

	if cfg.URL != "http://127.0.0.1:10080" {
		t.Errorf("expected 'http://127.0.0.1:10080', got %q", cfg.URL)
	}
}

func TestMissingAPIKey(t *testing.T) {
	_, err := parseConfig(t)
	if err == nil {
		t.Fatal("expected error due to missing required API key, got nil")
	}
}

func TestFlagsOverEnvVars(t *testing.T) {
	t.Setenv("TS_EXPORTER_API_KEY", "env-api-key")
	t.Setenv("TS_EXPORTER_LISTEN", ":9300")
	t.Setenv("TS_EXPORTER_URL", "http://192.168.0.100:10080")

	cfg, err := parseConfig(t, "--listen", ":9650", "--api-key", "flag-api-key")
	if err != nil {
		t.Fatalf("unexpected parsing error: %v", err)
	}

	if cfg.APIKey != "flag-api-key" {
		t.Errorf("expected 'flag-api-key', got %q", cfg.APIKey)
	}

	if cfg.Listen != ":9650" {
		t.Errorf("expected ':9650', got %q", cfg.Listen)
	}

	if cfg.URL != "http://192.168.0.100:10080" {
		t.Errorf("expected 'http://192.168.0.100:10080', got %q", cfg.URL)
	}
}
