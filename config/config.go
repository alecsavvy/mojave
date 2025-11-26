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
