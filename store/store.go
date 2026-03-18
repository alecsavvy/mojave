package store

import (
	"fmt"
	"iter"

	"github.com/cockroachdb/pebble"
	"google.golang.org/protobuf/proto"
)

type Store struct {
	*pebble.DB
}

func NewStore(db *pebble.DB) *Store {
	return &Store{
		DB: db,
	}
}

func (s *Store) GET(obj proto.Message, format string, args ...any) error {
	key := fmt.Appendf(nil, format, args...)
	value, closer, err := s.DB.Get(key)
	if err != nil {
		return err
	}
	defer closer.Close()

	return proto.Unmarshal(value, obj)
}

func (s *Store) SET(obj proto.Message, format string, args ...any) error {
	key := fmt.Appendf(nil, format, args...)
	value, err := proto.Marshal(obj)
	if err != nil {
		return err
	}
	return s.DB.Set(key, value, nil)
}

func (s *Store) BSET(batch *pebble.Batch, obj proto.Message, format string, args ...any) error {
	key := fmt.Appendf(nil, format, args...)
	value, err := proto.Marshal(obj)
	if err != nil {
		return err
	}
	return batch.Set(key, value, nil)
}

func (s *Store) ITER(def proto.Message, format string, args ...any) iter.Seq2[proto.Message, error] {
	return func(yield func(proto.Message, error) bool) {
		prefix := fmt.Appendf(nil, format, args...)
		it, err := s.DB.NewIter(&pebble.IterOptions{
			LowerBound: prefix,
			UpperBound: prefixUpperBound(prefix),
		})
		if err != nil {
			yield(nil, err)
			return
		}
		defer it.Close()

		for it.First(); it.Valid(); it.Next() {
			obj := proto.Clone(def)
			if err := proto.Unmarshal(it.Value(), obj); err != nil {
				yield(nil, err)
				return
			}
			if !yield(obj, nil) {
				return
			}
		}
	}
}

func (s *Store) DELETE(format string, args ...any) error {
	key := fmt.Appendf(nil, format, args...)
	return s.DB.Delete(key, nil)
}

func (s *Store) BDEL(batch *pebble.Batch, format string, args ...any) error {
	key := fmt.Appendf(nil, format, args...)
	return batch.Delete(key, nil)
}

func prefixUpperBound(prefix []byte) []byte {
	upper := make([]byte, len(prefix))
	copy(upper, prefix)
	for i := len(upper) - 1; i >= 0; i-- {
		upper[i]++
		if upper[i] != 0 {
			return upper[:i+1]
		}
	}
	return nil // prefix is all 0xff bytes, no upper bound needed
}
