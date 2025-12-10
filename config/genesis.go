package config

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/cometbft/cometbft/crypto"
	cmttypes "github.com/cometbft/cometbft/types"
)

// generates the genesis file for the Sonata chain
func GenerateGenesis(configDir string, validatorKeys []crypto.PubKey) error {
	validators := make([]cmttypes.GenesisValidator, len(validatorKeys))
	for i, validator := range validatorKeys {
		validators[i] = cmttypes.GenesisValidator{
			PubKey:  validator,
			Power:   10,
			Address: validator.Address(),
			Name:    fmt.Sprintf("validator-%d", i),
		}
	}

	chainID := fmt.Sprintf("mojave-%s", time.Now().Format("20060102150405"))
	genDoc := cmttypes.GenesisDoc{
		ChainID:     chainID,
		GenesisTime: time.Now(),
		Validators:  validators,
	}

	if err := genDoc.ValidateAndComplete(); err != nil {
		return fmt.Errorf("validating genesis doc: %w", err)
	}

	genesisFile := filepath.Join(configDir, "genesis.json")
	genDoc.SaveAs(genesisFile)

	return nil
}
