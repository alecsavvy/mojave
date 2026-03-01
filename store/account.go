package store

import (
	"context"
	"fmt"

	v1 "github.com/alecsavvy/mojave/gen/mojave/v1"
	"github.com/cockroachdb/pebble"
	"google.golang.org/protobuf/proto"
)

func accountKey(address string) []byte {
	return fmt.Appendf(nil, "account:%s", address)
}

func (s *Store) CreateAccount(ctx context.Context, batch *pebble.Batch, tx *v1.AccountState) error {
	key := accountKey(tx.Pubkey)

	value, err := proto.Marshal(tx)
	if err != nil {
		return err
	}

	return batch.Set(key, value, nil)
}

func (s *Store) GetAccount(ctx context.Context, address string) (*v1.AccountState, error) {
	key := accountKey(address)

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
