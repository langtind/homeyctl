package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Host   string `mapstructure:"host"`
	Port   int    `mapstructure:"port"`
	Token  string `mapstructure:"token"`
	Format string `mapstructure:"format"`
}

func (c *Config) BaseURL() string {
	return fmt.Sprintf("http://%s:%d", c.Host, c.Port)
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	// Config locations
	configDir, err := os.UserConfigDir()
	if err == nil {
		viper.AddConfigPath(filepath.Join(configDir, "homey-cli"))
	}
	viper.AddConfigPath(".")

	// Environment variables
	viper.SetEnvPrefix("HOMEY")
	viper.AutomaticEnv()

	// Defaults
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", 4859)
	viper.SetDefault("format", "json")

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

	dir := filepath.Join(configDir, "homey-cli")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	viper.Set("host", cfg.Host)
	viper.Set("port", cfg.Port)
	viper.Set("token", cfg.Token)
	viper.Set("format", cfg.Format)

	configPath := filepath.Join(dir, "config.toml")
	return viper.WriteConfigAs(configPath)
}
