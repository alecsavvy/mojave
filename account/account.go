package account

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
)

type AccountService struct {
	config *config.Config
}

var _ v1connect.AccountHandler = (*AccountService)(nil)

// Ping implements v1connect.AccountHandler.
func (a *AccountService) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	panic("unimplemented")
}

func NewAccountService(config *config.Config) *AccountService {
	return &AccountService{
		config: config,
	}
}
