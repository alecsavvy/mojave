package composition

import (
	"context"

	"connectrpc.com/connect"
	"github.com/alecsavvy/mojave/config"
	v1 "github.com/alecsavvy/mojave/gen/api/v1"
	"github.com/alecsavvy/mojave/gen/api/v1/v1connect"
	"github.com/alecsavvy/mojave/types/module"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	"go.uber.org/zap"
)

type CompositionService struct {
	*module.BaseModule
	config *config.Config
}

func (c *CompositionService) Name() string {
	return "composition"
}

var _ v1connect.CompositionHandler = (*CompositionService)(nil)

func (c *CompositionService) GetComposition(context.Context, *connect.Request[v1.GetCompositionRequest]) (*connect.Response[v1.GetCompositionResponse], error) {
	panic("unimplemented")
}

func NewCompositionService(config *config.Config, logger *zap.Logger) *CompositionService {
	svc := &CompositionService{config: config}
	svc.BaseModule = module.NewBaseModule(logger.Named(svc.Name()))
	return svc
}

// ABCI++ Callbacks

func (c *CompositionService) CheckTx(ctx context.Context, req *abcitypes.CheckTxRequest) (*abcitypes.CheckTxResponse, error) {
	c.Logger.Info("checking tx")
	return &abcitypes.CheckTxResponse{}, nil
}

func (c *CompositionService) PrepareProposal(ctx context.Context, req *abcitypes.PrepareProposalRequest) (*abcitypes.PrepareProposalResponse, error) {
	c.Logger.Info("preparing proposal")
	return &abcitypes.PrepareProposalResponse{}, nil
}

func (c *CompositionService) ProcessProposal(ctx context.Context, req *abcitypes.ProcessProposalRequest) (*abcitypes.ProcessProposalResponse, error) {
	c.Logger.Info("processing proposal")
	return &abcitypes.ProcessProposalResponse{}, nil
}

func (c *CompositionService) FinalizeBlock(ctx context.Context, req *abcitypes.FinalizeBlockRequest) (*abcitypes.FinalizeBlockResponse, error) {
	c.Logger.Info("finalizing block")
	return &abcitypes.FinalizeBlockResponse{}, nil
}
