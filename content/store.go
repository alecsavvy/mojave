package content

// Store is the content-addressed store (placeholder for blob/DAG storage).
type Store struct{}

// New returns a new Store.
func New() *Store {
	return &Store{}
}
