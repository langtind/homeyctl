package config

import (
	"testing"
)

func TestBaseURL(t *testing.T) {
	cfg := &Config{
		Host: "192.168.1.100",
		Port: 4859,
	}

	expected := "http://192.168.1.100:4859"
	if got := cfg.BaseURL(); got != expected {
		t.Errorf("BaseURL() = %q, want %q", got, expected)
	}
}

func TestBaseURLDefaultPort(t *testing.T) {
	cfg := &Config{
		Host: "localhost",
		Port: 80,
	}

	expected := "http://localhost:80"
	if got := cfg.BaseURL(); got != expected {
		t.Errorf("BaseURL() = %q, want %q", got, expected)
	}
}
