package app

import (
	"github.com/alecsavvy/mojave/config"
	"github.com/alecsavvy/mojave/content"
	"github.com/alecsavvy/mojave/store"
)

// App holds the application state and dependencies (ABCI app placeholder).
type App struct {
	cfg     *config.Config
	store   *store.Store
	content *content.Store
}

// New returns a new App.
func New(cfg *config.Config, st *store.Store, contentStore *content.Store) *App {
	return &App{
		cfg:     cfg,
		store:   st,
		content: contentStore,
	}
}
