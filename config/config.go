package config

import (
	"os"
	"path/filepath"

	"github.com/cometbft/cometbft/config"
)

const (
	DefaultHomeDir = ".sonata"
)

func DefaultHomeDirPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return DefaultHomeDir
	}
	return filepath.Join(home, DefaultHomeDir)
}

type Config struct {
	HTTP       *HTTPConfig
	Socket     *SocketConfig
	ChainStore *ChainStoreConfig
	LocalStore *LocalStoreConfig
	CometBFT   *config.Config
}

func DefaultConfig() *Config {
	return &Config{
		HTTP:       DefaultHTTPConfig(),
		Socket:     DefaultSocketConfig(),
		ChainStore: DefaultChainStoreConfig(),
		LocalStore: DefaultLocalStoreConfig(),
		CometBFT:   config.DefaultConfig(),
	}
}

type HTTPConfig struct {
	Host string
	Port int
}

func DefaultHTTPConfig() *HTTPConfig {
	return &HTTPConfig{
		Host: "0.0.0.0",
		Port: 8080,
	}
}

type SocketConfig struct {
	Path string
}

func DefaultSocketConfig() *SocketConfig {
	return &SocketConfig{
		Path: "unix:///tmp/sonata.sock",
	}
}

type ChainStoreConfig struct {
	Path string
}

func DefaultChainStoreConfig() *ChainStoreConfig {
	return &ChainStoreConfig{
		Path: filepath.Join(DefaultHomeDirPath(), "chainstore"),
	}
}

type LocalStoreConfig struct {
	Path string
}

func DefaultLocalStoreConfig() *LocalStoreConfig {
	return &LocalStoreConfig{
		Path: filepath.Join(DefaultHomeDirPath(), "localstore"),
	}
}
