package system

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
	"github.com/sonata-labs/sonata/types/module"
)

type SystemService struct {
	*module.BaseModule
	config *config.Config
}

func (s *SystemService) GetHealth(context.Context, *connect.Request[v1.GetHealthRequest]) (*connect.Response[v1.GetHealthResponse], error) {
	panic("unimplemented")
}

func (s *SystemService) GetNodeInfo(context.Context, *connect.Request[v1.GetNodeInfoRequest]) (*connect.Response[v1.GetNodeInfoResponse], error) {
	panic("unimplemented")
}

func (s *SystemService) GetReady(context.Context, *connect.Request[v1.GetReadyRequest]) (*connect.Response[v1.GetReadyResponse], error) {
	panic("unimplemented")
}

func (s *SystemService) GetStatus(context.Context, *connect.Request[v1.GetStatusRequest]) (*connect.Response[v1.GetStatusResponse], error) {
	panic("unimplemented")
}

var _ v1connect.SystemHandler = (*SystemService)(nil)

func NewSystemService(config *config.Config) *SystemService {
	return &SystemService{
		config: config,
	}
}
