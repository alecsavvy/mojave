package composition

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
	"github.com/sonata-labs/sonata/types/module"
)

type CompositionService struct {
	*module.BaseModule
	config *config.Config
}

var _ v1connect.CompositionHandler = (*CompositionService)(nil)

func (c *CompositionService) GetComposition(context.Context, *connect.Request[v1.GetCompositionRequest]) (*connect.Response[v1.GetCompositionResponse], error) {
	panic("unimplemented")
}

func NewCompositionService(config *config.Config) *CompositionService {
	return &CompositionService{
		config: config,
	}
}
