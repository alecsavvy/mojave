package account

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
	chainv1 "github.com/sonata-labs/sonata/gen/chain/v1"
	"github.com/sonata-labs/sonata/store/chainstore"
	"github.com/sonata-labs/sonata/types/module"
	"go.uber.org/zap"
)

type AccountService struct {
	*module.BaseModule
	store  *chainstore.ChainStore
	config *config.Config
}

func (a *AccountService) Name() string {
	return "account"
}

var _ module.Module = (*AccountService)(nil)

func (a *AccountService) GetAccount(ctx context.Context, req *connect.Request[v1.GetAccountRequest]) (*connect.Response[v1.GetAccountResponse], error) {
	address := req.Msg.Address
	if address == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("address is required"))
	}

	account, err := a.store.GetAccount(address)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	return connect.NewResponse(&v1.GetAccountResponse{Account: account}), nil
}

var _ v1connect.AccountHandler = (*AccountService)(nil)

func NewAccountService(config *config.Config, logger *zap.Logger, store *chainstore.ChainStore) *AccountService {
	svc := &AccountService{config: config, store: store}
	svc.BaseModule = module.NewBaseModule(logger.Named(svc.Name()))
	return svc
}

// ABCI++ Callbacks

func (a *AccountService) CheckTx(ctx context.Context, req *abcitypes.CheckTxRequest) (*abcitypes.CheckTxResponse, error) {
	a.Logger.Info("checking tx")
	return &abcitypes.CheckTxResponse{}, nil
}

func (a *AccountService) PrepareProposal(ctx context.Context, req *abcitypes.PrepareProposalRequest) (*abcitypes.PrepareProposalResponse, error) {
	a.Logger.Info("preparing proposal")
	return &abcitypes.PrepareProposalResponse{}, nil
}

func (a *AccountService) ProcessProposal(ctx context.Context, req *abcitypes.ProcessProposalRequest) (*abcitypes.ProcessProposalResponse, error) {
	a.Logger.Info("processing proposal")
	return &abcitypes.ProcessProposalResponse{}, nil
}

func (a *AccountService) FinalizeBlock(ctx context.Context, req *abcitypes.FinalizeBlockRequest) (*abcitypes.FinalizeBlockResponse, error) {
	a.Logger.Info("finalizing block")

	for _, tx := range req.Txs {
		var signedTransaction chainv1.SignedTransaction
		err := proto.Unmarshal(tx, &signedTransaction)
		if err != nil {
			return nil, err
		}

		if account := signedTransaction.Transaction.Body.GetCreateAccount(); account != nil {
			err = a.ChainStoreBatch.StoreAccount(account.Account)
			if err != nil {
				return nil, err
			}
			a.Logger.Info("created account", "address", account.Account.Address)
		}
	}

	return &abcitypes.FinalizeBlockResponse{}, nil
}
