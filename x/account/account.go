package account

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
	"github.com/sonata-labs/sonata/types/module"
	"go.uber.org/zap"
)

type AccountService struct {
	*module.BaseModule
	config *config.Config
}

func (a *AccountService) Name() string {
	return "account"
}

var _ module.Module = (*AccountService)(nil)

func (a *AccountService) GetAccount(context.Context, *connect.Request[v1.GetAccountRequest]) (*connect.Response[v1.GetAccountResponse], error) {
	panic("unimplemented")
}

var _ v1connect.AccountHandler = (*AccountService)(nil)

func NewAccountService(config *config.Config, logger *zap.Logger) *AccountService {
	svc := &AccountService{config: config}
	svc.BaseModule = module.NewBaseModule(logger.Named(svc.Name()))
	return svc
}
