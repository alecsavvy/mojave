package store

import (
	"github.com/cockroachdb/pebble"
)

// used to store local data that is not part of the blockchain
// stores the DEKs for files, keys, and stuff not meant to be replicated
type LocalStore struct {
	*pebble.DB
}

func NewLocalStore(db *pebble.DB) *LocalStore {
	return &LocalStore{
		DB: db,
	}
}
