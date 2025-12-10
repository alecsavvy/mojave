package module

import (
	"context"
	"sync"

	"github.com/alecsavvy/mojave/store/chainstore"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	"go.uber.org/zap"
)

type Module interface {
	// Name of the module
	Name() string

	SetChainStoreBatch(store *chainstore.ChainStore)

	// Startup dependency management
	RegisterStartupDeps(deps ...<-chan struct{})
	AwaitStartupDeps()
	Ready() <-chan struct{}

	// Shutdown dependency management
	RegisterShutdownDeps(deps ...<-chan struct{})
	AwaitShutdownDeps()
	Stopped() <-chan struct{}

	// Lifecycle methods
	Start() error
	Stop() error

	// ABCI++ methods
	abcitypes.Application
}

var _ Module = (*BaseModule)(nil)

type BaseModule struct {
	Logger          *zap.SugaredLogger
	ChainStoreBatch *chainstore.ChainStore
	ready           chan struct{}
	startupDeps     []<-chan struct{}
	stopped         chan struct{}
	shutdownDeps    []<-chan struct{}
}

func NewBaseModule(logger *zap.Logger) *BaseModule {
	return &BaseModule{
		Logger:  logger.Sugar(),
		ready:   make(chan struct{}),
		stopped: make(chan struct{}),
	}
}

func (m *BaseModule) Name() string {
	return ""
}

func (m *BaseModule) SetChainStoreBatch(store *chainstore.ChainStore) {
	m.ChainStoreBatch = store
}

func (m *BaseModule) RegisterStartupDeps(deps ...<-chan struct{}) {
	m.startupDeps = append(m.startupDeps, deps...)
}

func (m *BaseModule) Ready() <-chan struct{} {
	return m.ready
}

// MarkReady signals that the module is ready to handle requests
func (m *BaseModule) MarkReady() {
	close(m.ready)
}

// AwaitStartupDeps waits for all registered startup dependencies concurrently
func (m *BaseModule) AwaitStartupDeps() {
	var wg sync.WaitGroup
	wg.Add(len(m.startupDeps))
	for _, dep := range m.startupDeps {
		go func(ch <-chan struct{}) {
			<-ch
			wg.Done()
		}(dep)
	}
	wg.Wait()
}

func (m *BaseModule) Start() error {
	m.AwaitStartupDeps()
	m.Logger.Info("starting")
	m.MarkReady()
	return nil
}

func (m *BaseModule) RegisterShutdownDeps(deps ...<-chan struct{}) {
	m.shutdownDeps = append(m.shutdownDeps, deps...)
}

func (m *BaseModule) Stopped() <-chan struct{} {
	return m.stopped
}

// MarkStopped signals that the module has finished stopping
func (m *BaseModule) MarkStopped() {
	close(m.stopped)
}

// AwaitShutdownDeps waits for all registered shutdown dependencies concurrently
func (m *BaseModule) AwaitShutdownDeps() {
	var wg sync.WaitGroup
	wg.Add(len(m.shutdownDeps))
	for _, dep := range m.shutdownDeps {
		go func(ch <-chan struct{}) {
			<-ch
			wg.Done()
		}(dep)
	}
	wg.Wait()
}

func (m *BaseModule) Stop() error {
	m.AwaitShutdownDeps()
	m.Logger.Info("stopping")
	m.MarkStopped()
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
