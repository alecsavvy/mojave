package config

import (
	"fmt"

	"github.com/alecsavvy/mojave/common"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
)

// GenerateNodeKeys creates validator and node keys for a node.
// Returns the FilePV and NodeKey, or an error.
func GenerateNodeKeys(privValKeyFile, privValStateFile, nodeKeyFile string) (*privval.FilePV, *p2p.NodeKey, error) {
	var pv *privval.FilePV
	if common.FileExists(privValKeyFile) {
		pv = privval.LoadFilePV(privValKeyFile, privValStateFile)
	} else {
		genFilePV, err := privval.GenFilePV(privValKeyFile, privValStateFile, func() (crypto.PrivKey, error) {
			return secp256k1.GenPrivKey(), nil
		})
		if err != nil {
			return nil, nil, fmt.Errorf("gen file pv: %w", err)
		}
		genFilePV.Save()
		pv = privval.LoadFilePV(privValKeyFile, privValStateFile)
	}

	nodeKey, err := p2p.LoadOrGenNodeKey(nodeKeyFile)
	if err != nil {
		return nil, nil, fmt.Errorf("generate node key: %w", err)
	}

	return pv, nodeKey, nil
}
