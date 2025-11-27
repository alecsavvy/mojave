package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

func (app *App) runHTTP(_ context.Context) error {
	s := app.httpServer

	address := fmt.Sprintf("%s:%d", app.config.HTTP.Host, app.config.HTTP.Port)
	if err := s.Start(address); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
