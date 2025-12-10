package chain

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/alecsavvy/mojave/config"
	v1 "github.com/alecsavvy/mojave/gen/api/v1"
	"github.com/alecsavvy/mojave/gen/api/v1/v1connect"
	chainv1 "github.com/alecsavvy/mojave/gen/chain/v1"
	"github.com/alecsavvy/mojave/types/module"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/bytes"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/rpc/client/local"
	"github.com/cosmos/gogoproto/proto"
	"go.uber.org/zap"
)

var _ v1connect.ChainHandler = (*ChainService)(nil)

type ChainService struct {
	*module.BaseModule

	config *config.Config
	node   *node.Node
	rpc    *local.Local
}

func (c *ChainService) Name() string {
	return "chain"
}

func (c *ChainService) GetBlock(ctx context.Context, req *connect.Request[v1.GetBlockRequest]) (*connect.Response[v1.GetBlockResponse], error) {
	height := req.Msg.Height
	if height == 0 {
		height = c.node.BlockStore().Height()
	}

	block, blockMeta := c.node.BlockStore().LoadBlock(height)
	if block == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("block not found"))
	}

	protoBlock, err := block.ToProto()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoBlockMeta := blockMeta.ToProto()

	return connect.NewResponse(&v1.GetBlockResponse{Block: protoBlock, BlockMeta: protoBlockMeta}), nil
}

func (c *ChainService) GetTransaction(ctx context.Context, req *connect.Request[v1.GetTransactionRequest]) (*connect.Response[v1.GetTransactionResponse], error) {
	if req.Msg.TxHash == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("tx hash is required"))
	}

	txHash := bytes.HexBytes(req.Msg.TxHash)

	res, err := c.rpc.Tx(ctx, txHash, req.Msg.Prove)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	proof := res.Proof.ToProto()
	return connect.NewResponse(&v1.GetTransactionResponse{
		TxHash:   res.Hash.String(),
		Height:   res.Height,
		Index:    res.Index,
		TxResult: &res.TxResult,
		Tx:       res.Tx,
		Proof:    &proof,
	}), nil
}

func (c *ChainService) SendTransaction(ctx context.Context, req *connect.Request[v1.SendTransactionRequest]) (*connect.Response[v1.SendTransactionResponse], error) {
	rpc := c.rpc

	if len(req.Msg.SignedTransaction) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("signed transaction is required"))
	}

	res, err := rpc.BroadcastTxCommit(ctx, req.Msg.SignedTransaction)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.SendTransactionResponse{
		TxHash: res.Hash.String(),
	}), nil
}

func NewChainService(config *config.Config, logger *zap.Logger, node *node.Node) *ChainService {
	svc := &ChainService{config: config, node: node, rpc: local.New(node)}
	svc.BaseModule = module.NewBaseModule(logger.Named(svc.Name()))
	return svc
}

// ABCI++ Callbacks

func (c *ChainService) InitChain(ctx context.Context, req *abcitypes.InitChainRequest) (*abcitypes.InitChainResponse, error) {
	c.Logger.Info("initializing chain")
	return &abcitypes.InitChainResponse{}, nil
}

func (c *ChainService) CheckTx(ctx context.Context, req *abcitypes.CheckTxRequest) (*abcitypes.CheckTxResponse, error) {
	tx := req.Tx

	res := &abcitypes.CheckTxResponse{
		Code: 0,
	}

	var signedTransaction chainv1.SignedTransaction
	err := proto.Unmarshal(tx, &signedTransaction)
	if err != nil {
		res.Code = 1
		res.Info = "tx not a signed transaction"
		res.Log = err.Error()
		return res, nil
	}

	return res, nil
}

func (c *ChainService) PrepareProposal(ctx context.Context, req *abcitypes.PrepareProposalRequest) (*abcitypes.PrepareProposalResponse, error) {
	c.Logger.Info("preparing proposal")
	return &abcitypes.PrepareProposalResponse{Txs: req.Txs}, nil
}

func (c *ChainService) ProcessProposal(ctx context.Context, req *abcitypes.ProcessProposalRequest) (*abcitypes.ProcessProposalResponse, error) {
	c.Logger.Info("processing proposal")
	return &abcitypes.ProcessProposalResponse{Status: abcitypes.PROCESS_PROPOSAL_STATUS_ACCEPT}, nil
}

func (c *ChainService) FinalizeBlock(ctx context.Context, req *abcitypes.FinalizeBlockRequest) (*abcitypes.FinalizeBlockResponse, error) {
	c.Logger.Info("finalizing block")
	return &abcitypes.FinalizeBlockResponse{}, nil
}

func (c *ChainService) Commit(ctx context.Context, req *abcitypes.CommitRequest) (*abcitypes.CommitResponse, error) {
	c.Logger.Info("committing")
	return &abcitypes.CommitResponse{}, nil
}
