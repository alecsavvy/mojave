package app

import (
	mcrypto "github.com/alecsavvy/mojave/crypto"
	v1 "github.com/alecsavvy/mojave/gen/mojave/v1"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"google.golang.org/protobuf/proto"
)

func ResultTxToTransaction(resultTx *coretypes.ResultTx) (*v1.Transaction, *v1.TransactionResult, error) {
	transactionState, err := ResultTxToTransactionState(resultTx)
	if err != nil {
		return nil, nil, err
	}
	return TransactionStateToTransaction(transactionState)
}

func ResultTxToTransactionState(resultTx *coretypes.ResultTx) (*v1.TransactionState, error) {
	return &v1.TransactionState{
		SignedTransaction: resultTx.Tx,
		TransactionResult: resultTx.TxResult.Data,
	}, nil
}

func TransactionStateToTransaction(transactionState *v1.TransactionState) (*v1.Transaction, *v1.TransactionResult, error) {
	signedTransaction := &v1.SignedTransaction{}
	if err := proto.Unmarshal(transactionState.SignedTransaction, signedTransaction); err != nil {
		return nil, nil, err
	}

	transaction, err := mcrypto.VerifyTransaction(signedTransaction)
	if err != nil {
		return nil, nil, err
	}

	transactionResult := &v1.TransactionResult{}
	if err := proto.Unmarshal(transactionState.TransactionResult, transactionResult); err != nil {
		return nil, nil, err
	}

	return transaction, transactionResult, nil
}
