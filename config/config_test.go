package config

import (
	"testing"

	cfg "github.com/cometbft/cometbft/config"
)

func TestInitFilesWithConfig(t *testing.T) {
	config := cfg.DefaultConfig()
	config.SetRoot(t.TempDir())
	_, _, _, err := InitFilesWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to initialize files with config: %v", err)
	}
	t.Logf("Initialized CometBFT files under %s (inspect config/, genesis.json, etc.)", config.RootDir)
}
