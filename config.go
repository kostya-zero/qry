package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Borders string `koanf:"borders"`
}

func GetConfigPath() (string, error) {
	userDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(userDir, "qry", "config.toml")
	return configPath, nil
}

func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	if _, err = os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(filepath.Dir(configPath), 0o755)
		if err != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", err)
		}

		err = os.WriteFile(configPath, []byte{}, 0o644)
		if err != nil {
			return nil, fmt.Errorf("failed to create default config file: %w", err)
		}
	}

	k := koanf.New(".")

	k.Load(confmap.Provider(map[string]any{
		"borders": "rounded",
	}, "."), nil)

	if err := k.Load(file.Provider(configPath), toml.Parser()); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	var cfg Config

	k.Unmarshal("", &cfg)
	return &cfg, nil
}
