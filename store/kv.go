package store

import (
	"context"
	"fmt"

	v1 "github.com/alecsavvy/mojave/gen/mojave/v1"
	"github.com/cockroachdb/pebble"
	"google.golang.org/protobuf/proto"
)

func keyValueKey(key string) []byte {
	return fmt.Appendf(nil, "kv:%s", key)
}

func (s *Store) SetKeyValue(ctx context.Context, batch *pebble.Batch, tx *v1.KeyValueState) error {
	key := keyValueKey(tx.Key)

	value, err := proto.Marshal(tx)
	if err != nil {
		return err
	}

	return batch.Set(key, value, nil)
}

func (s *Store) GetKeyValue(ctx context.Context, k string) (*v1.KeyValueState, error) {
	key := keyValueKey(k)

	value, closer, err := s.DB.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	state := &v1.KeyValueState{}
	if err := proto.Unmarshal(value, state); err != nil {
		return nil, err
	}
	return state, nil
}
