package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecsavvy/mojave/config"
	"github.com/cometbft/cometbft/crypto"
	"github.com/spf13/cobra"
)

func NewInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize the Sonata node",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, _ := cmd.Flags().GetString("home")
			if home == "" {
				home = config.DefaultHomeDirPath()
			}

			// create home directory
			if err := os.MkdirAll(home, 0o755); err != nil {
				return fmt.Errorf("create dir %s: %w", home, err)
			}

			// create config directory
			configDir := filepath.Join(home, "config")
			if err := os.MkdirAll(configDir, 0o755); err != nil {
				return fmt.Errorf("create dir %s: %w", configDir, err)
			}

			// create data directory
			dataDir := filepath.Join(home, "data")
			if err := os.MkdirAll(dataDir, 0o755); err != nil {
				return fmt.Errorf("create dir %s: %w", dataDir, err)
			}

			// create default configuration
			cfg := config.DefaultConfig()
			cfg.SetRoot(home)

			// generate node keys (CometBFT expects key files in config/, state in data/)
			privValKeyFile := filepath.Join(configDir, "priv_validator_key.json")
			privValStateFile := filepath.Join(dataDir, "priv_validator_state.json")
			nodeKeyFile := filepath.Join(configDir, "node_key.json")
			pv, _, err := config.GenerateNodeKeys(privValKeyFile, privValStateFile, nodeKeyFile)
			if err != nil {
				return fmt.Errorf("generate node keys: %w", err)
			}

			// generate genesis file
			pubKey, err := pv.GetPubKey()
			if err != nil {
				return fmt.Errorf("get pub key: %w", err)
			}

			if err := config.GenerateGenesis(configDir, []crypto.PubKey{pubKey}); err != nil {
				return fmt.Errorf("generate genesis file: %w", err)
			}

			if err := config.WriteConfig(configDir, cfg); err != nil {
				return fmt.Errorf("write config: %w", err)
			}

			fmt.Println("Sonata node initialized successfully")
			return nil
		},
	}
}
