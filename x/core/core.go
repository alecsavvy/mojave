package core

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
)

var _ v1connect.CoreHandler = (*CoreService)(nil)

type CoreService struct {
	config *config.Config
}

// Ping implements v1connect.CoreHandler.
func (c *CoreService) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	panic("unimplemented")
}

func NewCoreService(config *config.Config) *CoreService {
	return &CoreService{
		config: config,
	}
}
