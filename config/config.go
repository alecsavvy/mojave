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

// GenValidatorKeys creates config/data dirs and generates priv val + node key only.
// It does not write config.toml or genesis. Used by testnet to get node IDs before setting persistent_peers.
func GenValidatorKeys(config *cfg.Config) (*privval.FilePV, *p2p.NodeKey, crypto.PubKey, error) {
	for _, subdir := range []string{"config", "data"} {
		if err := os.MkdirAll(filepath.Join(config.RootDir, subdir), 0755); err != nil {
			return nil, nil, nil, fmt.Errorf("create dir %s: %w", subdir, err)
		}
	}

	privValKeyFile := config.PrivValidatorKeyFile()
	privValStateFile := config.PrivValidatorStateFile()
	pv, err := privval.GenFilePV(privValKeyFile, privValStateFile, func() (crypto.PrivKey, error) {
		return ed25519.GenPrivKey(), nil
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("can't generate file pv: %w", err)
	}
	pv.Save()

	nodeKeyFile := config.NodeKeyFile()
	nodeKey, err := p2p.LoadOrGenNodeKey(nodeKeyFile)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("can't load or generate node key: %w", err)
	}

	pubKey, err := pv.GetPubKey()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("can't get pubkey: %w", err)
	}
	return pv, nodeKey, pubKey, nil
}

// InitValidatorOnly ensures dirs exist, writes config.toml, then generates priv val and node key.
// Caller must set P2P/RPC/persistent_peers on config before calling. Does not create genesis.
func InitValidatorOnly(config *cfg.Config) (*privval.FilePV, *p2p.NodeKey, crypto.PubKey, error) {
	for _, subdir := range []string{"config", "data"} {
		if err := os.MkdirAll(filepath.Join(config.RootDir, subdir), 0755); err != nil {
			return nil, nil, nil, fmt.Errorf("create dir %s: %w", subdir, err)
		}
	}
	cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
	return GenValidatorKeys(config)
}

// InitFilesWithConfig initializes a single-node directory: writes config, generates keys, and creates
// a single-validator genesis. Returns (pv, nodeKey, genDoc, nil).
func InitFilesWithConfig(config *cfg.Config) (*privval.FilePV, *p2p.NodeKey, *types.GenesisDoc, error) {
	pv, nodeKey, pubKey, err := InitValidatorOnly(config)
	if err != nil {
		return nil, nil, nil, err
	}

	genFile := config.GenesisFile()
	r := rand.Intn(1000000)
	genDoc := types.GenesisDoc{
		ChainID:         fmt.Sprintf("mojave-dev-%v", r),
		GenesisTime:     cmttime.Now(),
		ConsensusParams: types.DefaultConsensusParams(),
		Validators: []types.GenesisValidator{{
			Address: pubKey.Address(),
			PubKey:  pubKey,
			Power:   10,
		}},
	}
	if err := genDoc.SaveAs(genFile); err != nil {
		return nil, nil, nil, err
	}
	return pv, nodeKey, &genDoc, nil
}
