package storage

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
)

type StorageService struct {
	config *config.Config
}

var _ v1connect.StorageHandler = (*StorageService)(nil)

// Ping implements v1connect.StorageHandler.
func (s *StorageService) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	panic("unimplemented")
}

func NewStorageService(config *config.Config) *StorageService {
	return &StorageService{
		config: config,
	}
}
