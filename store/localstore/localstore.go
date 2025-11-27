package localstore

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
)

type LocalStoreService struct {
	config *config.Config
}

var _ v1connect.LocalStoreHandler = (*LocalStoreService)(nil)

// Ping implements v1connect.LocalStoreHandler.
func (l *LocalStoreService) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	panic("unimplemented")
}

func NewLocalStoreService(config *config.Config) *LocalStoreService {
	return &LocalStoreService{
		config: config,
	}
}
