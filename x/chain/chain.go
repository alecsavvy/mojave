package chain

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
	"github.com/sonata-labs/sonata/types/module"
	"go.uber.org/zap"
)

var _ v1connect.ChainHandler = (*ChainService)(nil)

type ChainService struct {
	*module.BaseModule
	config *config.Config
}

func (c *ChainService) Name() string {
	return "chain"
}

func (c *ChainService) GetBlock(context.Context, *connect.Request[v1.GetBlockRequest]) (*connect.Response[v1.GetBlockResponse], error) {
	panic("unimplemented")
}

func (c *ChainService) GetTransaction(context.Context, *connect.Request[v1.GetTransactionRequest]) (*connect.Response[v1.GetTransactionResponse], error) {
	panic("unimplemented")
}

func (c *ChainService) SendTransaction(context.Context, *connect.Request[v1.SendTransactionRequest]) (*connect.Response[v1.SendTransactionResponse], error) {
	panic("unimplemented")
}

func NewChainService(config *config.Config, logger *zap.Logger) *ChainService {
	svc := &ChainService{config: config}
	svc.BaseModule = module.NewBaseModule(logger.Named(svc.Name()))
	return svc
}
