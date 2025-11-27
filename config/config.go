package config

import (
	"os"
	"path/filepath"
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
	HTTP   *HTTPConfig
	Socket *SocketConfig
}

func DefaultConfig() *Config {
	return &Config{
		HTTP:   DefaultHTTPConfig(),
		Socket: DefaultSocketConfig(),
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
