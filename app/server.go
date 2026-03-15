package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"connectrpc.com/connect"
	mcrypto "github.com/alecsavvy/mojave/crypto"
	v1 "github.com/alecsavvy/mojave/gen/mojave/v1"
	"github.com/alecsavvy/mojave/gen/mojave/v1/v1connect"
	"github.com/alecsavvy/mojave/utils"
	"google.golang.org/protobuf/proto"
)

var _ v1connect.ServiceHandler = (*App)(nil)

func (app *App) GetKeyValue(ctx context.Context, req *connect.Request[v1.GetKeyValueRequest]) (*connect.Response[v1.GetKeyValueResponse], error) {
	kv, err := app.store.GetKeyValue(ctx, req.Msg.Key)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&v1.GetKeyValueResponse{KeyValue: kv}), nil
}

func (app *App) GetAccount(ctx context.Context, req *connect.Request[v1.GetAccountRequest]) (*connect.Response[v1.GetAccountResponse], error) {
	account, err := app.store.GetAccount(ctx, req.Msg.Pubkey)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&v1.GetAccountResponse{Account: account}), nil
}

func (app *App) GetTransaction(ctx context.Context, req *connect.Request[v1.GetTransactionRequest]) (*connect.Response[v1.GetTransactionResponse], error) {
	txHash := utils.HexToBytes(req.Msg.TxHash)
	tx, err := app.rpc.Tx(ctx, txHash, false)
	if err != nil {
		return nil, err
	}

	transaction, transactionResult, err := ResultTxToTransaction(tx)
	if err != nil {
		return nil, err
	}

	return &connect.Response[v1.GetTransactionResponse]{
		Msg: &v1.GetTransactionResponse{
			Transaction:       transaction,
			TransactionResult: transactionResult,
		},
	}, nil
}

func (app *App) SendTransaction(ctx context.Context, req *connect.Request[v1.SendTransactionRequest]) (*connect.Response[v1.SendTransactionResponse], error) {
	// ABCI expects the full marshaled SignedTransaction as the raw tx bytes.
	txBytes, err := proto.Marshal(req.Msg.SignedTransaction)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	res, err := app.rpc.BroadcastTxCommit(ctx, txBytes)
	if err != nil {
		return nil, err
	}

	// CheckTx failure: rejected from mempool, never went to consensus.
	if res.CheckTx.Code != 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("CheckTx failed (code %d): %s", res.CheckTx.Code, res.CheckTx.Log))
	}

	// ExecTxResult failure: made it into a block but the app rejected it.
	// Note: the tx IS on-chain at this point, just with a failed status.
	if res.TxResult.Code != 0 {
		return nil, connect.NewError(connect.CodeInternal,
			fmt.Errorf("ExecTx failed (code %d): %s", res.TxResult.Code, res.TxResult.Log))
	}

	txResult := &v1.TransactionResult{}
	if err := proto.Unmarshal(res.TxResult.Data, txResult); err != nil {
		return nil, connect.NewError(connect.CodeInternal,
			fmt.Errorf("failed to unmarshal tx result: %w", err))
	}

	return connect.NewResponse(&v1.SendTransactionResponse{
		TxHash:            res.Hash.String(),
		TransactionResult: txResult,
	}), nil
}

func (app *App) UploadFile(ctx context.Context, req *connect.Request[v1.UploadFileRequest]) (*connect.Response[v1.UploadFileResponse], error) {
	// validate signature
	hash := sha256.Sum256(req.Msg.FileData)
	if !mcrypto.VerifySignature(req.Msg.UploaderPubkey, req.Msg.Signature, hash[:]) {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("signature verification failed"))
	}

	// transcode file to flac
	// shell out to ffmpeg, transcode any audio file to flac
	cmd := exec.Command("ffmpeg", "-i", "pipe:0", "-f", "flac", "pipe:1")
	cmd.Stdin = bytes.NewReader(req.Msg.FileData)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	// stream flac bytes to local fs
	flacBytes := out.Bytes()
	os.WriteFile("flac.flac", flacBytes, 0644)

	// create DEK and encrypt file

	// store self dek in local store

	// create torrent file and get infohash

	// sign and send transaction to upload file
	// if tx fails then wipe and return error

	// store encrypted file in local fs and seed

	// return result

	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

func (app *App) GetFile(context.Context, *connect.Request[v1.GetFileRequest]) (*connect.Response[v1.GetFileResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}
