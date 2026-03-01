package store

import (
	"context"
	"fmt"

	v1 "github.com/alecsavvy/mojave/gen/mojave/v1"
	"github.com/cockroachdb/pebble"
	"google.golang.org/protobuf/proto"
)

func accountKey(pubkey []byte) []byte {
	return fmt.Appendf(nil, "account:%x", pubkey)
}

func (s *Store) GetAccount(ctx context.Context, pubkey []byte) (*v1.AccountState, error) {
	key := accountKey(pubkey)

	value, closer, err := s.DB.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	account := &v1.AccountState{}
	if err := proto.Unmarshal(value, account); err != nil {
		return nil, err
	}

	return account, nil
}

func (s *Store) GetOrCreateAccount(ctx context.Context, batch *pebble.Batch, pubkey []byte) (*v1.AccountState, error) {
	account, err := s.GetAccount(ctx, pubkey)
	if err != nil && err != pebble.ErrNotFound {
		return nil, err
	}

	account = &v1.AccountState{
		Pubkey:  pubkey,
		Balance: 0,
	}
	if err := s.UpdateAccount(ctx, batch, account); err != nil {
		return nil, err
	}
	return account, nil
}

// UpdateAccount takes an account state and updates the account in the batch.
func (s *Store) UpdateAccount(ctx context.Context, batch *pebble.Batch, tx *v1.AccountState) error {
	key := accountKey(tx.Pubkey)

	value, err := proto.Marshal(tx)
	if err != nil {
		return err
	}

	return batch.Set(key, value, nil)
}
