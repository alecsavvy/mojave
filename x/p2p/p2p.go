package p2p

import (
	"context"

	"connectrpc.com/connect"
	"github.com/alecsavvy/mojave/config"
	v1 "github.com/alecsavvy/mojave/gen/api/v1"
	"github.com/alecsavvy/mojave/gen/api/v1/v1connect"
	"github.com/alecsavvy/mojave/types/module"
	"go.uber.org/zap"
)

type P2PService struct {
	*module.BaseModule
	config *config.Config
}

func (p *P2PService) Name() string {
	return "p2p"
}

var _ v1connect.P2PHandler = (*P2PService)(nil)

func (p *P2PService) Stream(context.Context, *connect.BidiStream[v1.StreamRequest, v1.StreamResponse]) error {
	panic("unimplemented")
}

func NewP2PService(config *config.Config, logger *zap.Logger) *P2PService {
	svc := &P2PService{config: config}
	svc.BaseModule = module.NewBaseModule(logger.Named(svc.Name()))
	return svc
}
