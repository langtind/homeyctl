package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// LocalConfig holds settings for local (LAN/VPN) connection
type LocalConfig struct {
	Address string `mapstructure:"address"` // Full URL like http://192.168.1.50
	Token   string `mapstructure:"token"`   // Local API key from Homey web app
}

// CloudConfig holds settings for cloud connection
type CloudConfig struct {
	Token string `mapstructure:"token"` // Cloud token/PAT
}

type Config struct {
	// Legacy fields (still supported for backwards compatibility)
	Host   string `mapstructure:"host"`
	Port   int    `mapstructure:"port"`
	Token  string `mapstructure:"token"`
	Format string `mapstructure:"format"`
	TLS    bool   `mapstructure:"tls"`

	// New local/cloud mode fields
	Mode  string      `mapstructure:"mode"` // auto, local, cloud
	Local LocalConfig `mapstructure:"local"`
	Cloud CloudConfig `mapstructure:"cloud"`
}

// BaseURL returns the API base URL based on current mode
func (c *Config) BaseURL() string {
	mode := c.EffectiveMode()

	if mode == "local" {
		if c.Local.Address != "" {
			return c.Local.Address
		}
		// Fall back to legacy host/port if local address not set
		scheme := "http"
		if c.TLS {
			scheme = "https"
		}
		return fmt.Sprintf("%s://%s:%d", scheme, c.Host, c.Port)
	}

	// Cloud mode - use Homey cloud API
	// Note: Cloud API requires different handling, this is a placeholder
	return "https://api.athom.com"
}

// EffectiveMode returns the actual mode to use (resolves "auto")
func (c *Config) EffectiveMode() string {
	mode := c.Mode
	if mode == "" {
		mode = "auto"
	}

	if mode == "auto" {
		// Prefer local if address or legacy host is configured
		if c.Local.Address != "" || c.Host != "localhost" {
			return "local"
		}
		if c.Cloud.Token != "" {
			return "cloud"
		}
		// Default to local for backwards compatibility
		return "local"
	}

	return mode
}

// EffectiveToken returns the token for the current mode
func (c *Config) EffectiveToken() string {
	mode := c.EffectiveMode()

	if mode == "local" {
		if c.Local.Token != "" {
			return c.Local.Token
		}
		// Fall back to legacy token
		return c.Token
	}

	// Cloud mode
	if c.Cloud.Token != "" {
		return c.Cloud.Token
	}
	// Fall back to legacy token
	return c.Token
}

// CheckLegacyConfig checks if the old homey-cli config exists and prints migration instructions
func CheckLegacyConfig() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return
	}

	oldDir := filepath.Join(configDir, "homey-cli")
	newDir := filepath.Join(configDir, "homeyctl")

	// Check if old config exists and new one doesn't
	if _, err := os.Stat(filepath.Join(oldDir, "config.toml")); err == nil {
		if _, err := os.Stat(filepath.Join(newDir, "config.toml")); os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "⚠️  Found config from previous version (homey-cli)")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "The binary has been renamed from 'homey' to 'homeyctl' to avoid")
			fmt.Fprintln(os.Stderr, "conflicts with Athom's official Homey CLI for app development.")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "To migrate your config, run:")
			fmt.Fprintf(os.Stderr, "  mv %s %s\n", oldDir, newDir)
			fmt.Fprintln(os.Stderr, "")
		}
	}
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	// Config locations
	configDir, err := os.UserConfigDir()
	if err == nil {
		viper.AddConfigPath(filepath.Join(configDir, "homeyctl"))
	}
	viper.AddConfigPath(".")

	// Environment variables
	viper.SetEnvPrefix("HOMEY")
	viper.AutomaticEnv()

	// Explicitly bind env vars (required for keys without defaults)
	_ = viper.BindEnv("token")
	_ = viper.BindEnv("host")
	_ = viper.BindEnv("port")
	_ = viper.BindEnv("format")
	_ = viper.BindEnv("mode")
	_ = viper.BindEnv("address")       // HOMEY_ADDRESS for local mode
	_ = viper.BindEnv("local.token")   // HOMEY_LOCAL_TOKEN
	_ = viper.BindEnv("local.address") // HOMEY_LOCAL_ADDRESS

	// Defaults
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", 4859)
	viper.SetDefault("format", "json")
	viper.SetDefault("mode", "auto")

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config dir: %w", err)
	}

	dir := filepath.Join(configDir, "homeyctl")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	// Legacy fields
	viper.Set("host", cfg.Host)
	viper.Set("port", cfg.Port)
	viper.Set("token", cfg.Token)
	viper.Set("format", cfg.Format)
	viper.Set("tls", cfg.TLS)

	// New local/cloud mode fields
	viper.Set("mode", cfg.Mode)
	viper.Set("local.address", cfg.Local.Address)
	viper.Set("local.token", cfg.Local.Token)
	viper.Set("cloud.token", cfg.Cloud.Token)

	configPath := filepath.Join(dir, "config.toml")
	return viper.WriteConfigAs(configPath)
}
