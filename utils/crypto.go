package utils

import (
	"encoding/hex"

	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cometbft/cometbft/libs/bytes"
)

var ZeroAddress = make([]byte, 32)

// takes a byte array and returns the hex string
func HashHex(tx []byte) string {
	return bytes.HexBytes(tmhash.Sum(tx)).String()
}

// takes a hex string and returns the hash
func HexToBytes(tx string) []byte {
	hash, err := hex.DecodeString(tx)
	if err != nil {
		return []byte{}
	}
	return hash
}
