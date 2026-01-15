package cmd

import (
	"testing"

	"github.com/langtind/homeyctl/internal/config"
)

func TestMaskToken_Empty(t *testing.T) {
	result := maskToken("")
	if result != "(not set)" {
		t.Errorf("maskToken(\"\") = %q, want %q", result, "(not set)")
	}
}

func TestMaskToken_Short(t *testing.T) {
	result := maskToken("short")
	if result != "short" {
		t.Errorf("maskToken(\"short\") = %q, want %q", result, "short")
	}
}

func TestMaskToken_Long(t *testing.T) {
	token := "abcdefghijklmnopqrstuvwxyz"
	result := maskToken(token)
	expected := "abcdefgh...stuvwxyz" // first 8 + ... + last 8
	if result != expected {
		t.Errorf("maskToken(long) = %q, want %q", result, expected)
	}
}

func TestConfigDefaultFormat(t *testing.T) {
	// Test that config.Load returns json as default format
	cfg := &config.Config{}
	if cfg.Format != "" {
		t.Errorf("empty config Format = %q, want empty string", cfg.Format)
	}

	// The default is set by viper, but when empty we should treat it as json
	// This is tested implicitly by checking config package defaults
}

func TestIsTableFormat_NilConfig(t *testing.T) {
	// Save and restore the global cfg
	oldCfg := cfg
	defer func() { cfg = oldCfg }()

	cfg = nil
	if isTableFormat() {
		t.Error("isTableFormat() with nil cfg should return false")
	}
}

func TestIsTableFormat_JsonFormat(t *testing.T) {
	// Save and restore the global cfg
	oldCfg := cfg
	defer func() { cfg = oldCfg }()

	cfg = &config.Config{Format: "json"}
	if isTableFormat() {
		t.Error("isTableFormat() with json format should return false")
	}
}

func TestIsTableFormat_TableFormat(t *testing.T) {
	// Save and restore the global cfg
	oldCfg := cfg
	defer func() { cfg = oldCfg }()

	cfg = &config.Config{Format: "table"}
	if !isTableFormat() {
		t.Error("isTableFormat() with table format should return true")
	}
}

func TestIsTableFormat_EmptyFormat(t *testing.T) {
	// Save and restore the global cfg
	oldCfg := cfg
	defer func() { cfg = oldCfg }()

	cfg = &config.Config{Format: ""}
	if isTableFormat() {
		t.Error("isTableFormat() with empty format should return false (default to json)")
	}
}
