package store

// Store is the state store (placeholder for CometBFT/IAVL-backed state).
type Store struct{}

// New returns a new Store.
func New() *Store {
	return &Store{}
}
