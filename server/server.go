package server

import "github.com/alecsavvy/mojave/app"

// Server is the external interface (RPC, REST, gRPC placeholder).
type Server struct {
	app *app.App
}

// New returns a new Server.
func New(a *app.App) *Server {
	return &Server{app: a}
}

// Start starts the server. No-op for now.
func (s *Server) Start() error {
	return nil
}
