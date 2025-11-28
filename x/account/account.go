package account

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
	"github.com/sonata-labs/sonata/types/module"
)

type AccountService struct {
	*module.BaseModule
	config *config.Config
}

var _ module.Module = (*AccountService)(nil)

func (a *AccountService) GetAccount(context.Context, *connect.Request[v1.GetAccountRequest]) (*connect.Response[v1.GetAccountResponse], error) {
	panic("unimplemented")
}

var _ v1connect.AccountHandler = (*AccountService)(nil)

func NewAccountService(config *config.Config) *AccountService {
	return &AccountService{
		config: config,
	}
}
