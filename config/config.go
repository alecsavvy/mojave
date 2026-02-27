package config

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/types"
	cmttime "github.com/cometbft/cometbft/types/time"
)

func InitFilesWithConfig(config *cfg.Config) (*privval.FilePV, *p2p.NodeKey, *types.GenesisDoc, error) {
	// ensure dirs exist for config and other CometBFT files
	for _, subdir := range []string{"config", "data"} {
		if err := os.MkdirAll(filepath.Join(config.RootDir, subdir), 0755); err != nil {
			return nil, nil, nil, fmt.Errorf("create dir %s: %w", subdir, err)
		}
	}

	// write config file
	cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)

	// private validator
	privValKeyFile := config.PrivValidatorKeyFile()
	privValStateFile := config.PrivValidatorStateFile()
	var pv *privval.FilePV
	var err error
	pv, err = privval.GenFilePV(privValKeyFile, privValStateFile, func() (crypto.PrivKey, error) {
		return ed25519.GenPrivKey(), nil
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("can't generate file pv: %w", err)
	}
	pv.Save()

	nodeKeyFile := config.NodeKeyFile()
	if _, err := p2p.LoadOrGenNodeKey(nodeKeyFile); err != nil {
		return nil, nil, nil, err
	}

	nodeKey, err := p2p.LoadOrGenNodeKey(nodeKeyFile)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("can't load or generate node key: %w", err)
	}

	// genesis file
	genFile := config.GenesisFile()
	r := rand.Intn(1000000)
	genDoc := types.GenesisDoc{
		ChainID:         fmt.Sprintf("mojave-dev-%v", r),
		GenesisTime:     cmttime.Now(),
		ConsensusParams: types.DefaultConsensusParams(),
	}
	pubKey, err := pv.GetPubKey()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("can't get pubkey: %w", err)
	}
	genDoc.Validators = []types.GenesisValidator{{
		Address: pubKey.Address(),
		PubKey:  pubKey,
		Power:   10,
	}}

	if err := genDoc.SaveAs(genFile); err != nil {
		return nil, nil, nil, err
	}

	return pv, nodeKey, &genDoc, nil
}
