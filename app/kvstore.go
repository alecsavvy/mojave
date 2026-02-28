package app

import (
	"context"
	"fmt"

	v1 "github.com/alecsavvy/mojave/gen/mojave/v1"
	"github.com/alecsavvy/mojave/store"
	"github.com/cockroachdb/pebble"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type KVStoreApplication struct {
	logger       *zap.SugaredLogger
	store        *store.Store
	onGoingBlock *pebble.Batch
}

var _ abcitypes.Application = (*KVStoreApplication)(nil)

func NewKVStoreApplication(logger *zap.SugaredLogger, db *store.Store) *KVStoreApplication {
	return &KVStoreApplication{
		logger:       logger,
		store:        db,
		onGoingBlock: nil,
	}
}

func (app *KVStoreApplication) Info(_ context.Context, info *abcitypes.InfoRequest) (*abcitypes.InfoResponse, error) {
	return &abcitypes.InfoResponse{}, nil
}

func (app *KVStoreApplication) Query(ctx context.Context, req *abcitypes.QueryRequest) (*abcitypes.QueryResponse, error) {
	query := &v1.Query{}
	if err := proto.Unmarshal(req.Data, query); err != nil {
		return nil, err
	}

	switch query.Query.(type) {
	case *v1.Query_KeyValue:
		kvQuery := query.GetKeyValue()
		kv, err := app.store.GetKeyValue(ctx, kvQuery.Key)
		if err != nil {
			return nil, err
		}

		queryResponse := &v1.QueryResponse{
			Response: &v1.QueryResponse_KeyValue{
				KeyValue: kv,
			},
		}
		queryResponseBytes, err := proto.Marshal(queryResponse)
		if err != nil {
			return nil, err
		}
		return &abcitypes.QueryResponse{Value: queryResponseBytes}, nil
	}

	return nil, fmt.Errorf("unknown query type: %T", query.Query)
}

func (app *KVStoreApplication) CheckTx(_ context.Context, check *abcitypes.CheckTxRequest) (*abcitypes.CheckTxResponse, error) {
	return &abcitypes.CheckTxResponse{Code: 0}, nil
}

func (app *KVStoreApplication) InitChain(_ context.Context, chain *abcitypes.InitChainRequest) (*abcitypes.InitChainResponse, error) {
	return &abcitypes.InitChainResponse{}, nil
}

func (app *KVStoreApplication) PrepareProposal(_ context.Context, proposal *abcitypes.PrepareProposalRequest) (*abcitypes.PrepareProposalResponse, error) {
	return &abcitypes.PrepareProposalResponse{Txs: proposal.Txs}, nil
}

func (app *KVStoreApplication) ProcessProposal(_ context.Context, proposal *abcitypes.ProcessProposalRequest) (*abcitypes.ProcessProposalResponse, error) {
	return &abcitypes.ProcessProposalResponse{Status: abcitypes.PROCESS_PROPOSAL_STATUS_ACCEPT}, nil
}

func (app *KVStoreApplication) FinalizeBlock(_ context.Context, req *abcitypes.FinalizeBlockRequest) (*abcitypes.FinalizeBlockResponse, error) {
	var txs = make([]*abcitypes.ExecTxResult, len(req.Txs))

	app.onGoingBlock = app.store.NewBatch()
	for i, tx := range req.Txs {
		events, err := func(tx []byte) ([]abcitypes.Event, error) {
			var signedTransaction v1.SignedTransaction
			if err := proto.Unmarshal(tx, &signedTransaction); err != nil {
				return nil, err
			}

			// TODO: Verify signature
			var transaction v1.Transaction
			if err := proto.Unmarshal(signedTransaction.Transaction, &transaction); err != nil {
				return nil, err
			}

			switch transaction.Body.Body.(type) {
			case *v1.TransactionBody_CreateAccount:
				createAccountTx := transaction.Body.GetCreateAccount()
				account := &v1.AccountState{
					Address: createAccountTx.Pubkey,
				}
				if err := app.store.CreateAccount(context.Background(), app.onGoingBlock, account); err != nil {
					return nil, err
				}
				return []abcitypes.Event{}, nil
			case *v1.TransactionBody_KeyValue:
				kvTx := transaction.Body.GetKeyValue()
				kv := &v1.KeyValueState{
					Key:   kvTx.Key,
					Value: kvTx.Value,
				}
				if err := app.store.SetKeyValue(context.Background(), app.onGoingBlock, kv); err != nil {
					return nil, err
				}
				return []abcitypes.Event{}, nil
			default:
				return nil, fmt.Errorf("unknown transaction body type: %T", transaction.Body)
			}
		}(tx)

		code := abcitypes.CodeTypeOK
		if err != nil {
			app.logger.Errorf("Error processing transaction %v", err)
			code = 1
		}

		txs[i] = &abcitypes.ExecTxResult{
			Code:   code,
			Events: events,
		}
	}

	app.logger.Infow("finalized block", "height", req.Height, "txs", len(req.Txs))

	return &abcitypes.FinalizeBlockResponse{
		TxResults: txs,
	}, nil
}

func (app KVStoreApplication) Commit(_ context.Context, commit *abcitypes.CommitRequest) (*abcitypes.CommitResponse, error) {
	return &abcitypes.CommitResponse{}, app.onGoingBlock.Commit(nil)
}

func (app *KVStoreApplication) ListSnapshots(_ context.Context, snapshots *abcitypes.ListSnapshotsRequest) (*abcitypes.ListSnapshotsResponse, error) {
	return &abcitypes.ListSnapshotsResponse{}, nil
}

func (app *KVStoreApplication) OfferSnapshot(_ context.Context, snapshot *abcitypes.OfferSnapshotRequest) (*abcitypes.OfferSnapshotResponse, error) {
	return &abcitypes.OfferSnapshotResponse{}, nil
}

func (app *KVStoreApplication) LoadSnapshotChunk(_ context.Context, chunk *abcitypes.LoadSnapshotChunkRequest) (*abcitypes.LoadSnapshotChunkResponse, error) {
	return &abcitypes.LoadSnapshotChunkResponse{}, nil
}

func (app *KVStoreApplication) ApplySnapshotChunk(_ context.Context, chunk *abcitypes.ApplySnapshotChunkRequest) (*abcitypes.ApplySnapshotChunkResponse, error) {
	return &abcitypes.ApplySnapshotChunkResponse{Result: abcitypes.APPLY_SNAPSHOT_CHUNK_RESULT_ACCEPT}, nil
}

func (app KVStoreApplication) ExtendVote(_ context.Context, extend *abcitypes.ExtendVoteRequest) (*abcitypes.ExtendVoteResponse, error) {
	return &abcitypes.ExtendVoteResponse{}, nil
}

func (app *KVStoreApplication) VerifyVoteExtension(_ context.Context, verify *abcitypes.VerifyVoteExtensionRequest) (*abcitypes.VerifyVoteExtensionResponse, error) {
	return &abcitypes.VerifyVoteExtensionResponse{}, nil
}
