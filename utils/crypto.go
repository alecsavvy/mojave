package utils

import (
	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cometbft/cometbft/libs/bytes"
)

var ZeroAddress = make([]byte, 32)

func Hash(tx []byte) string {
	return bytes.HexBytes(tmhash.Sum(tx)).String()
}
