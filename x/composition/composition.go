package composition

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
	"github.com/sonata-labs/sonata/types/module"
	"go.uber.org/zap"
)

type CompositionService struct {
	*module.BaseModule
	config *config.Config
}

func (c *CompositionService) Name() string {
	return "composition"
}

var _ v1connect.CompositionHandler = (*CompositionService)(nil)

func (c *CompositionService) GetComposition(context.Context, *connect.Request[v1.GetCompositionRequest]) (*connect.Response[v1.GetCompositionResponse], error) {
	panic("unimplemented")
}

func NewCompositionService(config *config.Config, logger *zap.Logger) *CompositionService {
	svc := &CompositionService{config: config}
	svc.BaseModule = module.NewBaseModule(logger.Named(svc.Name()))
	return svc
}
