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

func (c *Config) ValidateBasic() error {
	return nil
}

func (c *Config) SetRoot(root string) {
	c.HTTP.Root = root
	c.Socket.Root = root
	c.ChainStore.Root = root
	c.LocalStore.Root = root
}

type HTTPConfig struct {
	Root string
	Host string
	Port int
}

func DefaultHTTPConfig() *HTTPConfig {
	return &HTTPConfig{
		Root: DefaultHomeDirPath(),
		Host: "0.0.0.0",
		Port: 8080,
	}
}

type SocketConfig struct {
	Root string
	Path string
}

func DefaultSocketConfig() *SocketConfig {
	return &SocketConfig{
		Root: DefaultHomeDirPath(),
		Path: "unix:///tmp/sonata.sock",
	}
}

type ChainStoreConfig struct {
	Root string
	Path string
}

func DefaultChainStoreConfig() *ChainStoreConfig {
	return &ChainStoreConfig{
		Root: DefaultHomeDirPath(),
		Path: filepath.Join(DefaultHomeDirPath(), "chainstore"),
	}
}

type LocalStoreConfig struct {
	Root string
	Path string
}

func DefaultLocalStoreConfig() *LocalStoreConfig {
	return &LocalStoreConfig{
		Root: DefaultHomeDirPath(),
		Path: filepath.Join(DefaultHomeDirPath(), "localstore"),
	}
}
