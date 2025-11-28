package core

import (
	"context"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/node"
	"github.com/sonata-labs/sonata/config"
	"github.com/sonata-labs/sonata/types/module"
)

type Core struct {
	config  *config.Config
	modules []module.Module
	node    *node.Node
}

func NewCore(config *config.Config, init func(c *Core) (*node.Node, error), modules ...module.Module) (*Core, *node.Node, error) {
	c := &Core{
		config:  config,
		modules: modules,
	}

	node, err := init(c)
	if err != nil {
		return nil, nil, err
	}
	c.node = node
	return c, node, nil
}

var _ module.Module = (*Core)(nil)

func (c *Core) Name() string {
	return "core"
}

func (c *Core) Start() error {
	if err := c.node.Start(); err != nil {
		return err
	}

	c.node.Wait()

	return nil
}

func (c *Core) Stop() error {
	if c.node != nil {
		return c.node.Stop()
	}
	return nil
}

// Info/Query Connection

func (c *Core) Info(ctx context.Context, req *abcitypes.InfoRequest) (*abcitypes.InfoResponse, error) {
	var lastResp *abcitypes.InfoResponse
	for _, mod := range c.modules {
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
	for _, mod := range c.modules {
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
	for _, mod := range c.modules {
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

	for _, mod := range c.modules {
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

	for _, mod := range c.modules {
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
	for _, mod := range c.modules {
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
	var txResults []*abcitypes.ExecTxResult
	var validators []abcitypes.ValidatorUpdate
	var events []abcitypes.Event
	var appHash []byte

	for _, mod := range c.modules {
		resp, err := mod.FinalizeBlock(ctx, req)
		if err != nil {
			return nil, err
		}
		if resp != nil {
			if len(resp.TxResults) > 0 {
				txResults = resp.TxResults
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

	for _, mod := range c.modules {
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
	for _, mod := range c.modules {
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

	for _, mod := range c.modules {
		resp, err := mod.Commit(ctx, req)
		if err != nil {
			return nil, err
		}
		if resp != nil && resp.RetainHeight > retainHeight {
			retainHeight = resp.RetainHeight
		}
	}

	return &abcitypes.CommitResponse{RetainHeight: retainHeight}, nil
}

// State Sync

func (c *Core) ListSnapshots(ctx context.Context, req *abcitypes.ListSnapshotsRequest) (*abcitypes.ListSnapshotsResponse, error) {
	var snapshots []*abcitypes.Snapshot

	for _, mod := range c.modules {
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
	for _, mod := range c.modules {
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
	for _, mod := range c.modules {
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
	for _, mod := range c.modules {
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
