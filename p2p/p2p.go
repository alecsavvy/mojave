package p2p

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
)

type P2PService struct {
	config *config.Config
}

var _ v1connect.P2PHandler = (*P2PService)(nil)

// Ping implements v1connect.P2PHandler.
func (p *P2PService) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	panic("unimplemented")
}

func NewP2PService(config *config.Config) *P2PService {
	return &P2PService{
		config: config,
	}
}
