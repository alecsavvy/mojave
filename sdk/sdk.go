package sdk

import (
	"context"
	"crypto/ed25519"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	mcrypto "github.com/alecsavvy/mojave/crypto"
	v1 "github.com/alecsavvy/mojave/gen/mojave/v1"
	"github.com/alecsavvy/mojave/gen/mojave/v1/v1connect"
	"github.com/alecsavvy/mojave/utils"
)

type MojaveSDK struct {
	privateKey ed25519.PrivateKey
	client     v1connect.ServiceClient
}

func NewMojaveSDK(connectURL string) (*MojaveSDK, error) {
	client := v1connect.NewServiceClient(http.DefaultClient, connectURL)
	return &MojaveSDK{client: client}, nil
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
	res, err := sdk.client.GetKeyValue(ctx, connect.NewRequest(&v1.GetKeyValueRequest{Key: key}))
	if err != nil {
		return nil, err
	}
	return res.Msg.KeyValue, nil
}

func (sdk *MojaveSDK) GetAccount(ctx context.Context, pubkey []byte) (*v1.AccountState, error) {
	res, err := sdk.client.GetAccount(ctx, connect.NewRequest(&v1.GetAccountRequest{Pubkey: pubkey}))
	if err != nil {
		return nil, err
	}
	return res.Msg.Account, nil
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

func (sdk *MojaveSDK) UploadFile(ctx context.Context, req *v1.UploadFileRequest) (*v1.FileUploadResult, error) {
	res, err := sdk.client.UploadFile(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return res.Msg.FileUploadResult, nil
}

func (sdk *MojaveSDK) GetFile(ctx context.Context, infohash string) (*v1.GetFileResponse, error) {
	res, err := sdk.client.GetFile(ctx, connect.NewRequest(&v1.GetFileRequest{Infohash: infohash}))
	if err != nil {
		return nil, err
	}
	return res.Msg, nil
}

func (sdk *MojaveSDK) sendTransaction(ctx context.Context, transaction *v1.SignedTransaction) (*v1.TransactionResult, error) {
	res, err := sdk.client.SendTransaction(ctx, connect.NewRequest(&v1.SendTransactionRequest{
		SignedTransaction: transaction,
	}))
	if err != nil {
		return nil, err
	}

	result := res.Msg.TransactionResult
	if result.Error != nil && result.Error.Code != v1.TransactionResultErrorCode_TRANSACTION_RESULT_ERROR_CODE_UNSPECIFIED {
		return result, errors.New(result.Error.Log)
	}

	return result, nil
}
