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

func TestBaseURLWithTLS(t *testing.T) {
	cfg := &Config{
		Host: "10-0-1-1.homey.homeylocal.com",
		Port: 4860,
		TLS:  true,
	}

	expected := "https://10-0-1-1.homey.homeylocal.com:4860"
	if got := cfg.BaseURL(); got != expected {
		t.Errorf("BaseURL() = %q, want %q", got, expected)
	}
}

func TestBaseURLWithoutTLS(t *testing.T) {
	cfg := &Config{
		Host: "10.0.1.1",
		Port: 4859,
		TLS:  false,
	}

	expected := "http://10.0.1.1:4859"
	if got := cfg.BaseURL(); got != expected {
		t.Errorf("BaseURL() = %q, want %q", got, expected)
	}
}

func TestEffectiveModeAuto(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		expected string
	}{
		{
			name:     "empty config defaults to local",
			cfg:      Config{Host: "localhost"},
			expected: "local",
		},
		{
			name: "auto with local address prefers local",
			cfg: Config{
				Mode:  "auto",
				Local: LocalConfig{Address: "http://192.168.1.50"},
			},
			expected: "local",
		},
		{
			name: "auto with only cloud token uses cloud",
			cfg: Config{
				Mode:  "auto",
				Host:  "localhost",
				Cloud: CloudConfig{Token: "cloud-token"},
			},
			expected: "cloud",
		},
		{
			name: "auto with both prefers local",
			cfg: Config{
				Mode:  "auto",
				Local: LocalConfig{Address: "http://192.168.1.50", Token: "local"},
				Cloud: CloudConfig{Token: "cloud"},
			},
			expected: "local",
		},
		{
			name: "explicit local mode",
			cfg: Config{
				Mode: "local",
			},
			expected: "local",
		},
		{
			name: "explicit cloud mode",
			cfg: Config{
				Mode: "cloud",
			},
			expected: "cloud",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.EffectiveMode(); got != tt.expected {
				t.Errorf("EffectiveMode() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestEffectiveToken(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		expected string
	}{
		{
			name: "local mode uses local token",
			cfg: Config{
				Mode:  "local",
				Local: LocalConfig{Token: "local-token"},
				Cloud: CloudConfig{Token: "cloud-token"},
				Token: "legacy-token",
			},
			expected: "local-token",
		},
		{
			name: "local mode falls back to legacy token",
			cfg: Config{
				Mode:  "local",
				Token: "legacy-token",
			},
			expected: "legacy-token",
		},
		{
			name: "cloud mode uses cloud token",
			cfg: Config{
				Mode:  "cloud",
				Local: LocalConfig{Token: "local-token"},
				Cloud: CloudConfig{Token: "cloud-token"},
				Token: "legacy-token",
			},
			expected: "cloud-token",
		},
		{
			name: "cloud mode falls back to legacy token",
			cfg: Config{
				Mode:  "cloud",
				Token: "legacy-token",
			},
			expected: "legacy-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.EffectiveToken(); got != tt.expected {
				t.Errorf("EffectiveToken() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestBaseURLWithLocalMode(t *testing.T) {
	cfg := &Config{
		Mode: "local",
		Local: LocalConfig{
			Address: "http://192.168.1.50",
		},
		Host: "legacy-host",
		Port: 4859,
	}

	expected := "http://192.168.1.50"
	if got := cfg.BaseURL(); got != expected {
		t.Errorf("BaseURL() = %q, want %q", got, expected)
	}
}

func TestBaseURLLocalFallsBackToLegacy(t *testing.T) {
	cfg := &Config{
		Mode: "local",
		Host: "192.168.1.100",
		Port: 4859,
	}

	expected := "http://192.168.1.100:4859"
	if got := cfg.BaseURL(); got != expected {
		t.Errorf("BaseURL() = %q, want %q", got, expected)
	}
}
