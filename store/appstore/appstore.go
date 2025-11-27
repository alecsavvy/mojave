package appstore

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
)

type AppStoreService struct {
	config *config.Config
}

var _ v1connect.AppStoreHandler = (*AppStoreService)(nil)

// Ping implements v1connect.AppStoreHandler.
func (a *AppStoreService) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	panic("unimplemented")
}

func NewAppStoreService(config *config.Config) *AppStoreService {
	return &AppStoreService{
		config: config,
	}
}
