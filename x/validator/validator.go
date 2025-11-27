package validator

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
)

type ValidatorService struct {
	config *config.Config
}

var _ v1connect.ValidatorHandler = (*ValidatorService)(nil)

// Ping implements v1connect.ValidatorHandler.
func (v *ValidatorService) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	panic("unimplemented")
}

func NewValidatorService(config *config.Config) *ValidatorService {
	return &ValidatorService{
		config: config,
	}
}
