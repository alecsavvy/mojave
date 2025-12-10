package statesync

import (
	"context"

	"github.com/alecsavvy/mojave/config"
	"github.com/alecsavvy/mojave/store/chainstore"
	"github.com/alecsavvy/mojave/types/module"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	"go.uber.org/zap"
)

// StateSyncService is a service that provides state sync functionality
type StateSyncService struct {
	*module.BaseModule

	config *config.Config
	logger *zap.Logger

	chainStore *chainstore.ChainStore
}

func NewStateSyncService(config *config.Config, logger *zap.Logger, chainStore *chainstore.ChainStore) *StateSyncService {
	svc := &StateSyncService{
		config:     config,
		logger:     logger,
		chainStore: chainStore,
	}
	svc.BaseModule = module.NewBaseModule(logger.Named(svc.Name()))
	return svc
}

func (s *StateSyncService) Name() string {
	return "statesync"
}

// ABCI++ Callbacks

func (s *StateSyncService) FinalizeBlock(ctx context.Context, req *abcitypes.FinalizeBlockRequest) (*abcitypes.FinalizeBlockResponse, error) {
	s.Logger.Info("finalizing block")
	return &abcitypes.FinalizeBlockResponse{}, nil
}

func (s *StateSyncService) Commit(ctx context.Context, req *abcitypes.CommitRequest) (*abcitypes.CommitResponse, error) {
	s.Logger.Info("committing")
	return &abcitypes.CommitResponse{}, nil
}

func (s *StateSyncService) ListSnapshots(ctx context.Context, req *abcitypes.ListSnapshotsRequest) (*abcitypes.ListSnapshotsResponse, error) {
	s.Logger.Info("listing snapshots")
	return &abcitypes.ListSnapshotsResponse{}, nil
}

func (s *StateSyncService) OfferSnapshot(ctx context.Context, req *abcitypes.OfferSnapshotRequest) (*abcitypes.OfferSnapshotResponse, error) {
	s.Logger.Info("offering snapshot")
	return &abcitypes.OfferSnapshotResponse{}, nil
}

func (s *StateSyncService) LoadSnapshotChunk(ctx context.Context, req *abcitypes.LoadSnapshotChunkRequest) (*abcitypes.LoadSnapshotChunkResponse, error) {
	s.Logger.Info("loading snapshot chunk")
	return &abcitypes.LoadSnapshotChunkResponse{}, nil
}

func (s *StateSyncService) ApplySnapshotChunk(ctx context.Context, req *abcitypes.ApplySnapshotChunkRequest) (*abcitypes.ApplySnapshotChunkResponse, error) {
	s.Logger.Info("applying snapshot chunk")
	return &abcitypes.ApplySnapshotChunkResponse{}, nil
}
