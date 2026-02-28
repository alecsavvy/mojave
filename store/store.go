package store

import (
	"github.com/cockroachdb/pebble"
)

type Store struct {
	*pebble.DB
}

func NewStore(db *pebble.DB) *Store {
	return &Store{
		DB: db,
	}
}
