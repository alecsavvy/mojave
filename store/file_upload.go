package store

import (
	"context"
	"fmt"

	v1 "github.com/alecsavvy/mojave/gen/mojave/v1"
	"github.com/cockroachdb/pebble"
	"google.golang.org/protobuf/proto"
)

func fileUploadKey(infohash string) []byte {
	return fmt.Appendf(nil, "file:%s", infohash)
}

func (s *Store) SetFileUpload(ctx context.Context, batch *pebble.Batch, record *v1.FileUploadTransaction) error {
	key := fileUploadKey(record.Infohash)
	value, err := proto.Marshal(record)
	if err != nil {
		return err
	}
	return batch.Set(key, value, nil)
}

func (s *Store) GetFileUpload(ctx context.Context, infohash string) (*v1.FileUploadTransaction, error) {
	key := fileUploadKey(infohash)
	value, closer, err := s.DB.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()
	record := &v1.FileUploadTransaction{}
	if err := proto.Unmarshal(value, record); err != nil {
		return nil, err
	}
	return record, nil
}
