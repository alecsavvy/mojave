package system

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
)

type SystemService struct {
	config *config.Config
}

var _ v1connect.SystemHandler = (*SystemService)(nil)

// Ping implements v1connect.SystemHandler.
func (s *SystemService) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	panic("unimplemented")
}

func NewSystemService(config *config.Config) *SystemService {
	return &SystemService{
		config: config,
	}
}
