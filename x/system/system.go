package system

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

type SystemService struct {
	*module.BaseModule
	config *config.Config
}

func (s *SystemService) Name() string {
	return "system"
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

func NewSystemService(config *config.Config, logger *zap.Logger) *SystemService {
	svc := &SystemService{config: config}
	svc.BaseModule = module.NewBaseModule(logger.Named(svc.Name()))
	return svc
}

// ABCI++ Callbacks

func (s *SystemService) PrepareProposal(ctx context.Context, req *abcitypes.PrepareProposalRequest) (*abcitypes.PrepareProposalResponse, error) {
	s.Logger.Info("preparing proposal")
	return &abcitypes.PrepareProposalResponse{}, nil
}

func (s *SystemService) ProcessProposal(ctx context.Context, req *abcitypes.ProcessProposalRequest) (*abcitypes.ProcessProposalResponse, error) {
	s.Logger.Info("processing proposal")
	return &abcitypes.ProcessProposalResponse{}, nil
}

func (s *SystemService) FinalizeBlock(ctx context.Context, req *abcitypes.FinalizeBlockRequest) (*abcitypes.FinalizeBlockResponse, error) {
	s.Logger.Info("finalizing block")
	return &abcitypes.FinalizeBlockResponse{}, nil
}
