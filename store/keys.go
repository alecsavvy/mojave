package store

import (
	"crypto/ed25519"
	"fmt"
)

func signingKey(pubkey []byte) []byte {
	return fmt.Appendf(nil, "signing:%x", pubkey)
}

func (s *LocalStore) SetSigningKey(pubkey []byte, privateKey ed25519.PrivateKey) error {
	key := signingKey(pubkey)
	return s.DB.Set(key, []byte(privateKey), nil)
}

func (s *LocalStore) GetSigningKey(pubkey []byte) (*ed25519.PrivateKey, error) {
	key := signingKey(pubkey)

	value, closer, err := s.DB.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	copied := make([]byte, len(value))
	copy(copied, value)

	pk := ed25519.PrivateKey(copied)
	return &pk, nil
}

func encryptionKey(pubkey []byte) []byte {
	return fmt.Appendf(nil, "encryption:%x", pubkey)
}

func (s *LocalStore) SetEncryptionKey(pubkey []byte, privateKey []byte) error {
	key := encryptionKey(pubkey)
	return s.DB.Set(key, privateKey, nil)
}

func (s *LocalStore) GetEncryptionKey(pubkey []byte) ([]byte, error) {
	key := encryptionKey(pubkey)

	value, closer, err := s.DB.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	copied := make([]byte, len(value))
	copy(copied, value)

	return copied, nil
}
