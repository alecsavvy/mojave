package config

import (
	"os"
	"path/filepath"

	"github.com/cometbft/cometbft/config"
	"github.com/pelletier/go-toml/v2"
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
	Sonata   *SonataConfig
	CometBFT *config.Config
}

func DefaultConfig() *Config {
	return &Config{
		Sonata:   DefaultSonataConfig(),
		CometBFT: config.DefaultConfig(),
	}
}

func (c *Config) ValidateBasic() error {
	return nil
}

func (c *Config) SetRoot(root string) {
	c.Sonata.SetRoot(root)
	c.CometBFT.SetRoot(root)
}

type SonataConfig struct {
	Root             string            `mapstructure:"root" toml:"root"`
	ChainID          string            `mapstructure:"chain_id" toml:"chain_id"`
	ValidatorAddress string            `mapstructure:"validator_address" toml:"validator_address"`
	HTTP             *HTTPConfig       `mapstructure:"http" toml:"http"`
	Socket           *SocketConfig     `mapstructure:"socket" toml:"socket"`
	ChainStore       *ChainStoreConfig `mapstructure:"chainstore" toml:"chainstore"`
	LocalStore       *LocalStoreConfig `mapstructure:"localstore" toml:"localstore"`
}

func DefaultSonataConfig() *SonataConfig {
	return &SonataConfig{
		Root:             DefaultHomeDirPath(),
		ChainID:          "sonata-1",
		ValidatorAddress: "",
		HTTP:             DefaultHTTPConfig(),
		Socket:           DefaultSocketConfig(),
		ChainStore:       DefaultChainStoreConfig(),
		LocalStore:       DefaultLocalStoreConfig(),
	}
}

func (c *SonataConfig) ValidateBasic() error {
	return nil
}

func (c *SonataConfig) SetRoot(root string) {
	c.Root = root
	c.HTTP.SetRoot(root)
	c.Socket.SetRoot(root)
	c.ChainStore.SetRoot(root)
	c.LocalStore.SetRoot(root)
}

type HTTPConfig struct {
	Root string `mapstructure:"root" toml:"root"`
	Host string `mapstructure:"host" toml:"host"`
	Port int    `mapstructure:"port" toml:"port"`
}

func DefaultHTTPConfig() *HTTPConfig {
	return &HTTPConfig{
		Root: DefaultHomeDirPath(),
		Host: "0.0.0.0",
		Port: 8080,
	}
}

func (c *HTTPConfig) SetRoot(root string) {
	c.Root = root
}

type SocketConfig struct {
	Root string `mapstructure:"root" toml:"root"`
	Path string `mapstructure:"path" toml:"path"`
}

func DefaultSocketConfig() *SocketConfig {
	return &SocketConfig{
		Root: DefaultHomeDirPath(),
		Path: "unix:///tmp/sonata.sock",
	}
}

func (c *SocketConfig) SetRoot(root string) {
	c.Root = root
}

type ChainStoreConfig struct {
	Root string `mapstructure:"root" toml:"root"`
	Path string `mapstructure:"path" toml:"path"`
}

func DefaultChainStoreConfig() *ChainStoreConfig {
	return &ChainStoreConfig{
		Root: DefaultHomeDirPath(),
		Path: filepath.Join(DefaultHomeDirPath(), "data", "chainstore.db"),
	}
}

func (c *ChainStoreConfig) SetRoot(root string) {
	c.Root = root
	c.Path = filepath.Join(root, "data", "chainstore.db")
}

type LocalStoreConfig struct {
	Root string `mapstructure:"root" toml:"root"`
	Path string `mapstructure:"path" toml:"path"`
}

func DefaultLocalStoreConfig() *LocalStoreConfig {
	return &LocalStoreConfig{
		Root: DefaultHomeDirPath(),
		Path: filepath.Join(DefaultHomeDirPath(), "data", "local.db"),
	}
}

func (c *LocalStoreConfig) SetRoot(root string) {
	c.Root = root
	c.Path = filepath.Join(root, "data", "local.db")
}

// SaveAs writes the SonataConfig to the specified file path as TOML.
func (c *SonataConfig) SaveAs(filePath string) error {
	data, err := toml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0o644)
}
