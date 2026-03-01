package sdk

import (
	"context"
	"crypto/ed25519"
	"errors"

	"github.com/cometbft/cometbft/rpc/client/http"
	"google.golang.org/protobuf/proto"

	mcrypto "github.com/alecsavvy/mojave/crypto"
	v1 "github.com/alecsavvy/mojave/gen/mojave/v1"
	"github.com/alecsavvy/mojave/utils"
)

type MojaveSDK struct {
	privateKey ed25519.PrivateKey
	*http.HTTP
}

func NewMojaveSDK(rpcURL string) (*MojaveSDK, error) {
	rpcClient, err := http.New(rpcURL)
	if err != nil {
		return nil, err
	}
	return &MojaveSDK{
		HTTP: rpcClient,
	}, nil
}

func (sdk *MojaveSDK) SetPrivateKey(privateKey ed25519.PrivateKey) {
	sdk.privateKey = privateKey
}

func (sdk *MojaveSDK) GetPublicKey() ed25519.PublicKey {
	return sdk.privateKey.Public().(ed25519.PublicKey)
}

func (sdk *MojaveSDK) SignTransaction(transaction *v1.Transaction) (*v1.SignedTransaction, error) {
	if sdk.privateKey == nil {
		return nil, errors.New("private key not set")
	}

	signedTransaction, err := mcrypto.SignTransaction(sdk.privateKey, transaction)
	if err != nil {
		return nil, err
	}
	return signedTransaction, nil
}

func (sdk *MojaveSDK) SetKeyValue(ctx context.Context, key string, value string) (*v1.KeyValueResult, error) {
	transaction := &v1.Transaction{
		Header: &v1.TransactionHeader{
			FromPubkey: sdk.GetPublicKey(),
		},
		Body: &v1.TransactionBody{
			Body: &v1.TransactionBody_KeyValue{
				KeyValue: &v1.KeyValueTransaction{Key: key, Value: value},
			},
		},
	}

	signedTransaction, err := sdk.SignTransaction(transaction)
	if err != nil {
		return nil, err
	}

	result, err := sdk.sendTransaction(ctx, signedTransaction)
	if err != nil {
		return nil, err
	}

	return result.Body.GetKeyValue(), nil
}

func (sdk *MojaveSDK) GetKeyValue(ctx context.Context, key string) (*v1.KeyValueState, error) {
	query := &v1.Query{
		Query: &v1.Query_KeyValue{
			KeyValue: &v1.KeyValueQuery{Key: key},
		},
	}

	response, err := sdk.sendQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	return response.GetKeyValue(), nil
}

func (sdk *MojaveSDK) GetAccount(ctx context.Context, pubkey []byte) (*v1.AccountState, error) {
	query := &v1.Query{
		Query: &v1.Query_Account{
			Account: &v1.AccountStateQuery{Pubkey: pubkey},
		},
	}

	response, err := sdk.sendQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	return response.GetAccount(), nil
}

func (sdk *MojaveSDK) TransferTokens(ctx context.Context, fromPubkey []byte, toPubkey []byte, amount uint64) (*v1.TokenTransferResult, error) {
	transaction := &v1.Transaction{
		Header: &v1.TransactionHeader{
			FromPubkey: sdk.GetPublicKey(),
		},

		Body: &v1.TransactionBody{
			Body: &v1.TransactionBody_TokenTransfer{
				TokenTransfer: &v1.TokenTransferTransaction{
					FromPubkey: fromPubkey,
					ToPubkey:   toPubkey,
					Amount:     amount,
				},
			},
		},
	}

	signedTransaction, err := sdk.SignTransaction(transaction)
	if err != nil {
		return nil, err
	}

	result, err := sdk.sendTransaction(ctx, signedTransaction)
	if err != nil {
		return nil, err
	}

	return result.Body.GetTokenTransfer(), nil
}

func (sdk *MojaveSDK) FaucetTokens(ctx context.Context, toPubkey []byte, amount uint64) error {
	_, err := sdk.TransferTokens(ctx, utils.ZeroAddress, toPubkey, amount)
	return err
}

func (sdk *MojaveSDK) sendTransaction(ctx context.Context, transaction *v1.SignedTransaction) (*v1.TransactionResult, error) {
	txBytes, err := proto.Marshal(transaction)
	if err != nil {
		return nil, err
	}

	response, err := sdk.HTTP.BroadcastTxCommit(ctx, txBytes)
	if err != nil {
		return nil, err
	}

	resultBytes := response.TxResult.Data

	result := &v1.TransactionResult{}
	if err := proto.Unmarshal(resultBytes, result); err != nil {
		return nil, err
	}

	resultCode := response.TxResult.Code

	if resultCode != 0 {
		return result, errors.New(result.Error.Log)
	}

	return result, nil
}

func (sdk *MojaveSDK) sendQuery(ctx context.Context, query *v1.Query) (*v1.QueryResponse, error) {
	queryBytes, err := proto.Marshal(query)
	if err != nil {
		return nil, err
	}
	response, err := sdk.HTTP.ABCIQuery(ctx, "", queryBytes)
	if err != nil {
		return nil, err
	}

	queryResponse := &v1.QueryResponse{}
	if err := proto.Unmarshal(response.Response.Value, queryResponse); err != nil {
		return nil, err
	}
	return queryResponse, nil
}
