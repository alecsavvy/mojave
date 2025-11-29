package module

import (
	"context"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"go.uber.org/zap"
)

type Module interface {
	// Name of the module
	Name() string

	// Lifecycle methods
	Start() error
	Stop() error

	// ABCI++ methods
	abcitypes.Application
}

var _ Module = (*BaseModule)(nil)

type BaseModule struct {
	Logger *zap.SugaredLogger
}

func NewBaseModule(logger *zap.Logger) *BaseModule {
	return &BaseModule{
		Logger: logger.Sugar(),
	}
}

func (m *BaseModule) Name() string {
	return ""
}

func (m *BaseModule) Start() error {
	m.Logger.Info("starting")
	return nil
}

func (m *BaseModule) Stop() error {
	m.Logger.Info("stopping")
	return nil
}

func (m *BaseModule) Info(ctx context.Context, req *abcitypes.InfoRequest) (*abcitypes.InfoResponse, error) {
	return &abcitypes.InfoResponse{}, nil
}

func (m *BaseModule) Query(ctx context.Context, req *abcitypes.QueryRequest) (*abcitypes.QueryResponse, error) {
	return &abcitypes.QueryResponse{}, nil
}

func (m *BaseModule) CheckTx(ctx context.Context, req *abcitypes.CheckTxRequest) (*abcitypes.CheckTxResponse, error) {
	return &abcitypes.CheckTxResponse{}, nil
}

func (m *BaseModule) InitChain(ctx context.Context, req *abcitypes.InitChainRequest) (*abcitypes.InitChainResponse, error) {
	return &abcitypes.InitChainResponse{}, nil
}

func (m *BaseModule) PrepareProposal(ctx context.Context, req *abcitypes.PrepareProposalRequest) (*abcitypes.PrepareProposalResponse, error) {
	return &abcitypes.PrepareProposalResponse{}, nil
}

func (m *BaseModule) ProcessProposal(ctx context.Context, req *abcitypes.ProcessProposalRequest) (*abcitypes.ProcessProposalResponse, error) {
	return &abcitypes.ProcessProposalResponse{}, nil
}

func (m *BaseModule) FinalizeBlock(ctx context.Context, req *abcitypes.FinalizeBlockRequest) (*abcitypes.FinalizeBlockResponse, error) {
	return &abcitypes.FinalizeBlockResponse{}, nil
}

func (m *BaseModule) ExtendVote(ctx context.Context, req *abcitypes.ExtendVoteRequest) (*abcitypes.ExtendVoteResponse, error) {
	return &abcitypes.ExtendVoteResponse{}, nil
}

func (m *BaseModule) VerifyVoteExtension(ctx context.Context, req *abcitypes.VerifyVoteExtensionRequest) (*abcitypes.VerifyVoteExtensionResponse, error) {
	return &abcitypes.VerifyVoteExtensionResponse{}, nil
}

func (m *BaseModule) Commit(ctx context.Context, req *abcitypes.CommitRequest) (*abcitypes.CommitResponse, error) {
	return &abcitypes.CommitResponse{}, nil
}

func (m *BaseModule) ListSnapshots(ctx context.Context, req *abcitypes.ListSnapshotsRequest) (*abcitypes.ListSnapshotsResponse, error) {
	return &abcitypes.ListSnapshotsResponse{}, nil
}

func (m *BaseModule) OfferSnapshot(ctx context.Context, req *abcitypes.OfferSnapshotRequest) (*abcitypes.OfferSnapshotResponse, error) {
	return &abcitypes.OfferSnapshotResponse{}, nil
}

func (m *BaseModule) LoadSnapshotChunk(ctx context.Context, req *abcitypes.LoadSnapshotChunkRequest) (*abcitypes.LoadSnapshotChunkResponse, error) {
	return &abcitypes.LoadSnapshotChunkResponse{}, nil
}

func (m *BaseModule) ApplySnapshotChunk(ctx context.Context, req *abcitypes.ApplySnapshotChunkRequest) (*abcitypes.ApplySnapshotChunkResponse, error) {
	return &abcitypes.ApplySnapshotChunkResponse{}, nil
}
