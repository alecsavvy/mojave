package chainstore

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
)

type ChainStoreService struct {
	config *config.Config
}

var _ v1connect.ChainStoreHandler = (*ChainStoreService)(nil)

// Ping implements v1connect.ChainStoreHandler.
func (a *ChainStoreService) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	panic("unimplemented")
}

func NewChainStoreService(config *config.Config) *ChainStoreService {
	return &ChainStoreService{
		config: config,
	}
}
