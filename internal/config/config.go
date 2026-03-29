package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Vault     string         `json:"vault"`
	Editor    string         `json:"editor"`
	NoOpen    bool           `json:"noOpen"`
	Defaults  DefaultsConfig `json:"defaults"`
	Templates string         `json:"templates"`
	Queries   string         `json:"queries"`
}

type DefaultsConfig struct {
	StaleDays      int `json:"staleDays"`
	DashboardLimit int `json:"dashboardLimit"`
}

func Load() (Config, error) {
	cfg := Config{
		Defaults: DefaultsConfig{
			StaleDays:      30,
			DashboardLimit: 5,
		},
	}

	configPath, err := path()
	if err != nil {
		return cfg, err
	}

	raw, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := json.Unmarshal(raw, &cfg); err != nil {
		return cfg, err
	}

	if cfg.Defaults.StaleDays == 0 {
		cfg.Defaults.StaleDays = 30
	}
	if cfg.Defaults.DashboardLimit == 0 {
		cfg.Defaults.DashboardLimit = 5
	}

	return cfg, nil
}

func Save(cfg Config) error {
	configPath, err := path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(configPath, raw, 0o644)
}

func Path() (string, error) {
	return path()
}

func path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".yanp", "config.json"), nil
}
