package core

import (
	"context"
	"sync"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/node"
	"github.com/sonata-labs/sonata/config"
	"github.com/sonata-labs/sonata/store/chainstore"
	"github.com/sonata-labs/sonata/types/module"
	"go.uber.org/zap"
)

// Callback represents an ABCI++ callback type
type Callback int

const (
	Info Callback = iota
	Query
	CheckTx
	InitChain
	PrepareProposal
	ProcessProposal
	FinalizeBlock
	ExtendVote
	VerifyVoteExtension
	Commit
	ListSnapshots
	OfferSnapshot
	LoadSnapshotChunk
	ApplySnapshotChunk
)

type Core struct {
	config  *config.Config
	modules map[Callback][]module.Module
	node    *node.Node
	logger  *zap.SugaredLogger

	ready        chan struct{}
	startupDeps  []<-chan struct{}
	stopped      chan struct{}
	shutdownDeps []<-chan struct{}

	chainStore *chainstore.ChainStore
	batch      *chainstore.ChainStore
}

func NewCore(config *config.Config, logger *zap.Logger, init func(c *Core) (*node.Node, error), chainStore *chainstore.ChainStore) (*Core, *node.Node, error) {
	c := &Core{
		config:     config,
		modules:    make(map[Callback][]module.Module),
		logger:     logger.Named("core").Sugar(),
		ready:      make(chan struct{}),
		stopped:    make(chan struct{}),
		chainStore: chainStore,
	}

	node, err := init(c)
	if err != nil {
		return nil, nil, err
	}
	c.node = node
	return c, node, nil
}

// RegisterModules registers modules for a specific ABCI callback.
// Modules are executed in the order they are provided.
func (c *Core) RegisterModules(callback Callback, modules ...module.Module) {
	c.modules[callback] = modules
}

var _ module.Module = (*Core)(nil)

func (c *Core) Name() string {
	return "core"
}

func (c *Core) SetChainStoreBatch(store *chainstore.ChainStore) {}

func (c *Core) RegisterStartupDeps(deps ...<-chan struct{}) {
	c.startupDeps = append(c.startupDeps, deps...)
}

func (c *Core) Ready() <-chan struct{} {
	return c.ready
}

// AwaitStartupDeps waits for all registered startup dependencies concurrently
func (c *Core) AwaitStartupDeps() {
	var wg sync.WaitGroup
	wg.Add(len(c.startupDeps))
	for _, dep := range c.startupDeps {
		go func(ch <-chan struct{}) {
			<-ch
			wg.Done()
		}(dep)
	}
	wg.Wait()
}

func (c *Core) RegisterShutdownDeps(deps ...<-chan struct{}) {
	c.shutdownDeps = append(c.shutdownDeps, deps...)
}

func (c *Core) Stopped() <-chan struct{} {
	return c.stopped
}

// AwaitShutdownDeps waits for all registered shutdown dependencies concurrently
func (c *Core) AwaitShutdownDeps() {
	var wg sync.WaitGroup
	wg.Add(len(c.shutdownDeps))
	for _, dep := range c.shutdownDeps {
		go func(ch <-chan struct{}) {
			<-ch
			wg.Done()
		}(dep)
	}
	wg.Wait()
}

func (c *Core) Start() error {
	c.AwaitStartupDeps()
	c.logger.Info("all modules ready, starting cometbft")

	if err := c.node.Start(); err != nil {
		return err
	}
	close(c.ready)

	c.node.Wait()

	return nil
}

func (c *Core) Stop() error {
	c.AwaitShutdownDeps()
	c.logger.Info("stopping")

	var err error
	if c.node != nil {
		err = c.node.Stop()
	}
	close(c.stopped)
	return err
}

// Info/Query Connection

func (c *Core) Info(ctx context.Context, req *abcitypes.InfoRequest) (*abcitypes.InfoResponse, error) {
	var lastResp *abcitypes.InfoResponse
	for _, mod := range c.modules[Info] {
		resp, err := mod.Info(ctx, req)
		if err != nil {
			return nil, err
		}
		lastResp = resp
	}
	if lastResp == nil {
		return &abcitypes.InfoResponse{}, nil
	}
	return lastResp, nil
}

func (c *Core) Query(ctx context.Context, req *abcitypes.QueryRequest) (*abcitypes.QueryResponse, error) {
	for _, mod := range c.modules[Query] {
		resp, err := mod.Query(ctx, req)
		if err != nil {
			return nil, err
		}
		// If a module handles the query (non-empty response), return it
		if resp != nil && len(resp.Value) > 0 {
			return resp, nil
		}
	}
	return &abcitypes.QueryResponse{}, nil
}

// Mempool Connection

func (c *Core) CheckTx(ctx context.Context, req *abcitypes.CheckTxRequest) (*abcitypes.CheckTxResponse, error) {
	for _, mod := range c.modules[CheckTx] {
		resp, err := mod.CheckTx(ctx, req)
		if err != nil {
			return nil, err
		}
		// If any module rejects the tx, return immediately
		if resp != nil && resp.Code != 0 {
			return resp, nil
		}
	}
	return &abcitypes.CheckTxResponse{Code: 0}, nil
}

// Consensus Connection

func (c *Core) InitChain(ctx context.Context, req *abcitypes.InitChainRequest) (*abcitypes.InitChainResponse, error) {
	var validators []abcitypes.ValidatorUpdate
	var appHash []byte

	for _, mod := range c.modules[InitChain] {
		resp, err := mod.InitChain(ctx, req)
		if err != nil {
			return nil, err
		}
		if resp != nil {
			if len(resp.Validators) > 0 {
				validators = resp.Validators
			}
			if len(resp.AppHash) > 0 {
				appHash = resp.AppHash
			}
		}
	}

	return &abcitypes.InitChainResponse{
		Validators: validators,
		AppHash:    appHash,
	}, nil
}

func (c *Core) PrepareProposal(ctx context.Context, req *abcitypes.PrepareProposalRequest) (*abcitypes.PrepareProposalResponse, error) {
	txs := req.Txs

	for _, mod := range c.modules[PrepareProposal] {
		resp, err := mod.PrepareProposal(ctx, req)
		if err != nil {
			return nil, err
		}
		if resp != nil && len(resp.Txs) > 0 {
			txs = resp.Txs
		}
	}

	return &abcitypes.PrepareProposalResponse{Txs: txs}, nil
}

func (c *Core) ProcessProposal(ctx context.Context, req *abcitypes.ProcessProposalRequest) (*abcitypes.ProcessProposalResponse, error) {
	for _, mod := range c.modules[ProcessProposal] {
		resp, err := mod.ProcessProposal(ctx, req)
		if err != nil {
			return nil, err
		}
		// If any module rejects the proposal, reject immediately
		if resp != nil && resp.Status == abcitypes.PROCESS_PROPOSAL_STATUS_REJECT {
			return resp, nil
		}
	}
	return &abcitypes.ProcessProposalResponse{Status: abcitypes.PROCESS_PROPOSAL_STATUS_ACCEPT}, nil
}

func (c *Core) FinalizeBlock(ctx context.Context, req *abcitypes.FinalizeBlockRequest) (*abcitypes.FinalizeBlockResponse, error) {
	var validators []abcitypes.ValidatorUpdate
	var events []abcitypes.Event
	var appHash []byte
	c.batch = c.chainStore.Batch()

	// Initialize tx results - one per transaction in the block
	txResults := make([]*abcitypes.ExecTxResult, len(req.Txs))
	for i := range txResults {
		txResults[i] = &abcitypes.ExecTxResult{Code: 0}
	}

	for _, mod := range c.modules[FinalizeBlock] {
		// set module chain store to the core controlled batch
		mod.SetChainStoreBatch(c.batch)
		resp, err := mod.FinalizeBlock(ctx, req)
		if err != nil {
			// Convert module error to failed tx results instead of failing consensus
			c.logger.Errorw("module error during FinalizeBlock", "module", mod.Name(), "error", err)
			for i := range txResults {
				if txResults[i].Code == 0 {
					txResults[i] = &abcitypes.ExecTxResult{
						Code: 1,
						Log:  err.Error(),
					}
				}
			}
			continue
		}
		if resp != nil {
			// Merge tx results from modules (if they provide any)
			for i, result := range resp.TxResults {
				if i < len(txResults) && result != nil {
					txResults[i] = result
				}
			}
			if len(resp.ValidatorUpdates) > 0 {
				validators = resp.ValidatorUpdates
			}
			if len(resp.Events) > 0 {
				events = append(events, resp.Events...)
			}
			if len(resp.AppHash) > 0 {
				appHash = resp.AppHash
			}
		}
	}

	return &abcitypes.FinalizeBlockResponse{
		TxResults:        txResults,
		ValidatorUpdates: validators,
		Events:           events,
		AppHash:          appHash,
	}, nil
}

// Vote Extensions

func (c *Core) ExtendVote(ctx context.Context, req *abcitypes.ExtendVoteRequest) (*abcitypes.ExtendVoteResponse, error) {
	var extension []byte

	for _, mod := range c.modules[ExtendVote] {
		resp, err := mod.ExtendVote(ctx, req)
		if err != nil {
			return nil, err
		}
		if resp != nil && len(resp.VoteExtension) > 0 {
			extension = append(extension, resp.VoteExtension...)
		}
	}

	return &abcitypes.ExtendVoteResponse{VoteExtension: extension}, nil
}

func (c *Core) VerifyVoteExtension(ctx context.Context, req *abcitypes.VerifyVoteExtensionRequest) (*abcitypes.VerifyVoteExtensionResponse, error) {
	for _, mod := range c.modules[VerifyVoteExtension] {
		resp, err := mod.VerifyVoteExtension(ctx, req)
		if err != nil {
			return nil, err
		}
		// If any module rejects the vote extension, reject immediately
		if resp != nil && resp.Status == abcitypes.VERIFY_VOTE_EXTENSION_STATUS_REJECT {
			return resp, nil
		}
	}
	return &abcitypes.VerifyVoteExtensionResponse{Status: abcitypes.VERIFY_VOTE_EXTENSION_STATUS_ACCEPT}, nil
}

// Commit

func (c *Core) Commit(ctx context.Context, req *abcitypes.CommitRequest) (*abcitypes.CommitResponse, error) {
	var retainHeight int64

	for _, mod := range c.modules[Commit] {
		resp, err := mod.Commit(ctx, req)
		if err != nil {
			return nil, err
		}
		if resp != nil && resp.RetainHeight > retainHeight {
			retainHeight = resp.RetainHeight
		}
		// once the commit is complete, set the chain store batch to nil
		mod.SetChainStoreBatch(nil)
	}

	// commit the batch from all modules
	if err := c.batch.Commit(); err != nil {
		return nil, err
	}

	return &abcitypes.CommitResponse{RetainHeight: retainHeight}, nil
}

// State Sync

func (c *Core) ListSnapshots(ctx context.Context, req *abcitypes.ListSnapshotsRequest) (*abcitypes.ListSnapshotsResponse, error) {
	var snapshots []*abcitypes.Snapshot

	for _, mod := range c.modules[ListSnapshots] {
		resp, err := mod.ListSnapshots(ctx, req)
		if err != nil {
			return nil, err
		}
		if resp != nil && len(resp.Snapshots) > 0 {
			snapshots = append(snapshots, resp.Snapshots...)
		}
	}

	return &abcitypes.ListSnapshotsResponse{Snapshots: snapshots}, nil
}

func (c *Core) OfferSnapshot(ctx context.Context, req *abcitypes.OfferSnapshotRequest) (*abcitypes.OfferSnapshotResponse, error) {
	for _, mod := range c.modules[OfferSnapshot] {
		resp, err := mod.OfferSnapshot(ctx, req)
		if err != nil {
			return nil, err
		}
		// If any module rejects the snapshot, return that result
		if resp != nil && resp.Result != abcitypes.OFFER_SNAPSHOT_RESULT_ACCEPT {
			return resp, nil
		}
	}
	return &abcitypes.OfferSnapshotResponse{Result: abcitypes.OFFER_SNAPSHOT_RESULT_ACCEPT}, nil
}

func (c *Core) LoadSnapshotChunk(ctx context.Context, req *abcitypes.LoadSnapshotChunkRequest) (*abcitypes.LoadSnapshotChunkResponse, error) {
	for _, mod := range c.modules[LoadSnapshotChunk] {
		resp, err := mod.LoadSnapshotChunk(ctx, req)
		if err != nil {
			return nil, err
		}
		// Return the first non-empty chunk
		if resp != nil && len(resp.Chunk) > 0 {
			return resp, nil
		}
	}
	return &abcitypes.LoadSnapshotChunkResponse{}, nil
}

func (c *Core) ApplySnapshotChunk(ctx context.Context, req *abcitypes.ApplySnapshotChunkRequest) (*abcitypes.ApplySnapshotChunkResponse, error) {
	for _, mod := range c.modules[ApplySnapshotChunk] {
		resp, err := mod.ApplySnapshotChunk(ctx, req)
		if err != nil {
			return nil, err
		}
		// If any module rejects or needs refetch, return that result
		if resp != nil && resp.Result != abcitypes.APPLY_SNAPSHOT_CHUNK_RESULT_ACCEPT {
			return resp, nil
		}
	}
	return &abcitypes.ApplySnapshotChunkResponse{Result: abcitypes.APPLY_SNAPSHOT_CHUNK_RESULT_ACCEPT}, nil
}
