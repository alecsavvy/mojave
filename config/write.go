package config

import (
	"fmt"
	"path/filepath"

	cmtconfig "github.com/cometbft/cometbft/config"
)

// WriteConfig writes the cometbft config.toml and mojave.toml files to the config directory
func WriteConfig(configDir string, cfg *Config) error {
	// write cometbft config.toml
	cmtconfig.WriteConfigFile(filepath.Join(configDir, "config.toml"), cfg.CometBFT)

	// write mojave.toml
	if err := cfg.Mojave.SaveAs(filepath.Join(configDir, "mojave.toml")); err != nil {
		return fmt.Errorf("writing mojave.toml: %w", err)
	}
	return nil
}
