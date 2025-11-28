package p2p

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
	"github.com/sonata-labs/sonata/types/module"
)

type P2PService struct {
	*module.BaseModule
	config *config.Config
}

var _ v1connect.P2PHandler = (*P2PService)(nil)

func (p *P2PService) Stream(context.Context, *connect.BidiStream[v1.StreamRequest, v1.StreamResponse]) error {
	panic("unimplemented")
}

func NewP2PService(config *config.Config) *P2PService {
	return &P2PService{
		config: config,
	}
}
