package sdk

import (
	"context"
	"fmt"

	"github.com/cometbft/cometbft/rpc/client/http"
	"google.golang.org/protobuf/proto"

	v1 "github.com/alecsavvy/mojave/gen/mojave/v1"
)

type MojaveSDK struct {
	rpcClient *http.HTTP
}

func NewMojaveSDK(rpcURL string) (*MojaveSDK, error) {
	rpcClient, err := http.New(rpcURL)
	if err != nil {
		return nil, err
	}
	return &MojaveSDK{
		rpcClient: rpcClient,
	}, nil
}

func (sdk *MojaveSDK) SetKeyValue(ctx context.Context, key string, value string) error {
	transaction := &v1.Transaction{
		Body: &v1.TransactionBody{
			Body: &v1.TransactionBody_KeyValue{
				KeyValue: &v1.KeyValueTransaction{Key: key, Value: value},
			},
		},
	}

	txBytes, err := proto.Marshal(transaction)
	if err != nil {
		return err
	}

	signedTransaction := &v1.SignedTransaction{
		Transaction: txBytes,
	}

	return sdk.sendTransaction(ctx, signedTransaction)
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

func (sdk *MojaveSDK) sendTransaction(ctx context.Context, transaction *v1.SignedTransaction) error {
	txBytes, err := proto.Marshal(transaction)
	if err != nil {
		return err
	}

	response, err := sdk.rpcClient.BroadcastTxCommit(ctx, txBytes)
	if err != nil {
		return err
	}
	if response.CheckTx.Code != 0 {
		return fmt.Errorf("transaction failed: %s", response.CheckTx.Log)
	}
	return nil
}

func (sdk *MojaveSDK) sendQuery(ctx context.Context, query *v1.Query) (*v1.QueryResponse, error) {
	queryBytes, err := proto.Marshal(query)
	if err != nil {
		return nil, err
	}
	response, err := sdk.rpcClient.ABCIQuery(ctx, "", queryBytes)
	if err != nil {
		return nil, err
	}

	queryResponse := &v1.QueryResponse{}
	if err := proto.Unmarshal(response.Response.Value, queryResponse); err != nil {
		return nil, err
	}
	return queryResponse, nil
}
