package validator

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

type ValidatorService struct {
	*module.BaseModule
	config *config.Config
}

func (v *ValidatorService) Name() string {
	return "validator"
}

func (v *ValidatorService) GetValidator(context.Context, *connect.Request[v1.GetValidatorRequest]) (*connect.Response[v1.GetValidatorResponse], error) {
	panic("unimplemented")
}

func (v *ValidatorService) GetValidators(context.Context, *connect.Request[v1.GetValidatorsRequest]) (*connect.Response[v1.GetValidatorsResponse], error) {
	panic("unimplemented")
}

var _ v1connect.ValidatorHandler = (*ValidatorService)(nil)

func NewValidatorService(config *config.Config, logger *zap.Logger) *ValidatorService {
	svc := &ValidatorService{config: config}
	svc.BaseModule = module.NewBaseModule(logger.Named(svc.Name()))
	return svc
}

// ABCI++ Callbacks

func (v *ValidatorService) CheckTx(ctx context.Context, req *abcitypes.CheckTxRequest) (*abcitypes.CheckTxResponse, error) {
	v.Logger.Info("checking tx")
	return &abcitypes.CheckTxResponse{}, nil
}

func (v *ValidatorService) PrepareProposal(ctx context.Context, req *abcitypes.PrepareProposalRequest) (*abcitypes.PrepareProposalResponse, error) {
	v.Logger.Info("preparing proposal")
	return &abcitypes.PrepareProposalResponse{}, nil
}

func (v *ValidatorService) ProcessProposal(ctx context.Context, req *abcitypes.ProcessProposalRequest) (*abcitypes.ProcessProposalResponse, error) {
	v.Logger.Info("processing proposal")
	return &abcitypes.ProcessProposalResponse{}, nil
}

func (v *ValidatorService) FinalizeBlock(ctx context.Context, req *abcitypes.FinalizeBlockRequest) (*abcitypes.FinalizeBlockResponse, error) {
	v.Logger.Info("finalizing block")
	return &abcitypes.FinalizeBlockResponse{}, nil
}
