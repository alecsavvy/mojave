package config

import (
	"fmt"
	"path/filepath"

	cmtconfig "github.com/cometbft/cometbft/config"
	"github.com/spf13/viper"
)

func ReadConfig(homeDir string) (*Config, error) {
	if homeDir == "" {
		homeDir = DefaultHomeDirPath()
	}

	// Read Sonata config
	mojaveConfig, err := ReadMojaveConfig(homeDir)
	if err != nil {
		return nil, fmt.Errorf("reading mojave config: %w", err)
	}

	// Read CometBFT config
	cmtConfig, err := ReadCometBFTConfig(homeDir)
	if err != nil {
		return nil, fmt.Errorf("reading cometbft config: %w", err)
	}

	return &Config{
		Mojave:   mojaveConfig,
		CometBFT: cmtConfig,
	}, nil
}

func ReadMojaveConfig(homeDir string) (*MojaveConfig, error) {
	mojaveConfig := DefaultMojaveConfig()
	mojaveConfig.SetRoot(homeDir)

	v := viper.New()
	v.SetConfigFile(filepath.Join(homeDir, "config", "mojave.toml"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	if err := v.Unmarshal(mojaveConfig); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}
	if err := mojaveConfig.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return mojaveConfig, nil
}

func ReadCometBFTConfig(homeDir string) (*cmtconfig.Config, error) {
	cmtConfig := cmtconfig.DefaultConfig()
	cmtConfig.SetRoot(homeDir)

	v := viper.New()
	v.SetConfigFile(filepath.Join(homeDir, "config", "config.toml"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	if err := v.Unmarshal(cmtConfig); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}
	if err := cmtConfig.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return cmtConfig, nil
}
