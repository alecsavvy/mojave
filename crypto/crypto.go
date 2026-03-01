package crypto

import (
	"crypto/ed25519"
	"crypto/sha256"
	"errors"

	v1 "github.com/alecsavvy/mojave/gen/mojave/v1"
	"github.com/cosmos/gogoproto/proto"
)

// SignTransaction signs a transaction with an Ed25519 private key.
// Returns the signed transaction and the raw tx bytes. Protobuf marshalling is not deterministic,
// so the same logical transaction can produce different bytes (and thus different signatures) across calls.
func SignTransaction(privateKey ed25519.PrivateKey, transaction *v1.Transaction) (*v1.SignedTransaction, error) {
	txBytes, err := proto.Marshal(transaction)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(txBytes)
	signature := ed25519.Sign(privateKey, hash[:])

	return &v1.SignedTransaction{
		Transaction: txBytes,
		Signature:   signature,
	}, nil
}

// VerifyTransaction verifies an Ed25519 signature over the transaction bytes and unmarshals the transaction.
func VerifyTransaction(signedTransaction *v1.SignedTransaction) (*v1.Transaction, error) {
	var transaction v1.Transaction
	if err := proto.Unmarshal(signedTransaction.Transaction, &transaction); err != nil {
		return nil, err
	}

	if transaction.Header == nil {
		return nil, errors.New("transaction header is nil")
	}

	if len(transaction.Header.FromPubkey) == 0 {
		return nil, errors.New("transaction from pubkey is empty")
	}

	publicKey := ed25519.PublicKey(transaction.Header.FromPubkey)
	hash := sha256.Sum256(signedTransaction.Transaction)
	if !ed25519.Verify(publicKey, hash[:], signedTransaction.Signature) {
		return nil, errors.New("signature verification failed")
	}

	return &transaction, nil
}
