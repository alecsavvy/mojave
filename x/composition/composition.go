package composition

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
)

type CompositionService struct {
	config *config.Config
}

var _ v1connect.CompositionHandler = (*CompositionService)(nil)

// Ping implements v1connect.CompositionHandler.
func (c *CompositionService) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	panic("unimplemented")
}

func NewCompositionService(config *config.Config) *CompositionService {
	return &CompositionService{
		config: config,
	}
}
