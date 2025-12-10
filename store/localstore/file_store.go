package localstore

import (
	"bytes"
	"errors"

	storagev1 "github.com/alecsavvy/mojave/gen/storage/v1"
	"github.com/cockroachdb/pebble"
	"google.golang.org/protobuf/proto"
)

// StoreUpload stores a file and its metadata in the local store.
func (l *LocalStore) StoreUpload(cid string, data []byte, meta *storagev1.UploadMeta) error {
	// Store the file data
	if err := l.db.Set(uploadKey(cid), data, pebble.Sync); err != nil {
		return err
	}

	// Store the metadata
	metaBytes, err := proto.Marshal(meta)
	if err != nil {
		return err
	}
	return l.db.Set(uploadMetaKey(cid), metaBytes, pebble.Sync)
}

// GetUpload retrieves a file from the local store.
func (l *LocalStore) GetUpload(cid string) ([]byte, error) {
	data, closer, err := l.db.Get(uploadKey(cid))
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	// Copy data since it's only valid until closer is closed
	result := make([]byte, len(data))
	copy(result, data)
	return result, nil
}

// GetUploadMeta retrieves upload metadata from the local store.
func (l *LocalStore) GetUploadMeta(cid string) (*storagev1.UploadMeta, error) {
	data, closer, err := l.db.Get(uploadMetaKey(cid))
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	meta := &storagev1.UploadMeta{}
	if err := proto.Unmarshal(data, meta); err != nil {
		return nil, err
	}
	return meta, nil
}

// StoreChunk stores a chunk of a file. Metadata is stored with the first chunk (index 0)
// or when provided for any chunk if metadata doesn't exist yet.
func (l *LocalStore) StoreChunk(cid string, index uint32, data []byte, meta *storagev1.UploadMeta) error {
	// Store the chunk
	if err := l.db.Set(chunkKey(cid, index), data, pebble.Sync); err != nil {
		return err
	}

	// Store metadata if provided and not already stored
	if meta != nil {
		_, closer, err := l.db.Get(uploadMetaKey(cid))
		if errors.Is(err, pebble.ErrNotFound) {
			metaBytes, err := proto.Marshal(meta)
			if err != nil {
				return err
			}
			return l.db.Set(uploadMetaKey(cid), metaBytes, pebble.Sync)
		} else if err != nil {
			return err
		}
		closer.Close()
	}

	return nil
}

// GetChunk retrieves a chunk from the local store.
func (l *LocalStore) GetChunk(cid string, index uint32) ([]byte, error) {
	data, closer, err := l.db.Get(chunkKey(cid, index))
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	result := make([]byte, len(data))
	copy(result, data)
	return result, nil
}

// HasAllChunks checks if all chunks for a file have been uploaded.
func (l *LocalStore) HasAllChunks(cid string) (bool, error) {
	meta, err := l.GetUploadMeta(cid)
	if err != nil {
		return false, err
	}

	for i := uint32(0); i < meta.TotalChunks; i++ {
		_, closer, err := l.db.Get(chunkKey(cid, i))
		if errors.Is(err, pebble.ErrNotFound) {
			return false, nil
		} else if err != nil {
			return false, err
		}
		closer.Close()
	}

	return true, nil
}

// CountChunks returns the number of chunks that have been uploaded.
func (l *LocalStore) CountChunks(cid string) (uint32, error) {
	meta, err := l.GetUploadMeta(cid)
	if err != nil {
		return 0, err
	}

	var count uint32
	for i := uint32(0); i < meta.TotalChunks; i++ {
		_, closer, err := l.db.Get(chunkKey(cid, i))
		if errors.Is(err, pebble.ErrNotFound) {
			continue
		} else if err != nil {
			return 0, err
		}
		closer.Close()
		count++
	}

	return count, nil
}

// ReassembleChunks concatenates all chunks in order and returns the full file.
func (l *LocalStore) ReassembleChunks(cid string) ([]byte, error) {
	meta, err := l.GetUploadMeta(cid)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	for i := uint32(0); i < meta.TotalChunks; i++ {
		chunk, err := l.GetChunk(cid, i)
		if err != nil {
			return nil, err
		}
		buf.Write(chunk)
	}

	return buf.Bytes(), nil
}

// DeleteChunks removes all chunks for a file.
func (l *LocalStore) DeleteChunks(cid string) error {
	meta, err := l.GetUploadMeta(cid)
	if err != nil {
		return err
	}

	for i := uint32(0); i < meta.TotalChunks; i++ {
		if err := l.db.Delete(chunkKey(cid, i), pebble.Sync); err != nil && !errors.Is(err, pebble.ErrNotFound) {
			return err
		}
	}

	return nil
}

// StoreTranscoded stores a transcoded file.
func (l *LocalStore) StoreTranscoded(cid string, data []byte) error {
	return l.db.Set(transcodedKey(cid), data, pebble.Sync)
}

// GetTranscoded retrieves a transcoded file.
func (l *LocalStore) GetTranscoded(cid string) ([]byte, error) {
	data, closer, err := l.db.Get(transcodedKey(cid))
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	result := make([]byte, len(data))
	copy(result, data)
	return result, nil
}

// DeleteUpload removes the original file and its metadata.
func (l *LocalStore) DeleteUpload(cid string) error {
	if err := l.db.Delete(uploadKey(cid), pebble.Sync); err != nil && !errors.Is(err, pebble.ErrNotFound) {
		return err
	}
	if err := l.db.Delete(uploadMetaKey(cid), pebble.Sync); err != nil && !errors.Is(err, pebble.ErrNotFound) {
		return err
	}
	return nil
}

// HasUpload checks if an upload exists.
func (l *LocalStore) HasUpload(cid string) bool {
	_, closer, err := l.db.Get(uploadKey(cid))
	if err != nil {
		return false
	}
	closer.Close()
	return true
}
