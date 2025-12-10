package storage

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/alecsavvy/mojave/common/cid"
	"github.com/alecsavvy/mojave/config"
	v1 "github.com/alecsavvy/mojave/gen/api/v1"
	"github.com/alecsavvy/mojave/gen/api/v1/v1connect"
	chainv1 "github.com/alecsavvy/mojave/gen/chain/v1"
	storagev1 "github.com/alecsavvy/mojave/gen/storage/v1"
	"github.com/alecsavvy/mojave/media"
	"github.com/alecsavvy/mojave/store/chainstore"
	"github.com/alecsavvy/mojave/store/localstore"
	"github.com/alecsavvy/mojave/types/module"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/gogoproto/proto"
	"go.uber.org/zap"
)

const (
	MaxDirectUploadSize = 50 * 1024 * 1024 // 50MB
	MaxChunkSize        = 10 * 1024 * 1024 // 10MB
	MaxEncoderWorkers   = 4
)

type StorageService struct {
	*module.BaseModule
	config     *config.Config
	localStore *localstore.LocalStore
	chainStore *chainstore.ChainStore
	encoder    *media.MediaEncoder
	chain      v1connect.ChainHandler
}

// SetChain sets the chain handler dependency (in-process, no network).
func (s *StorageService) SetChain(chain v1connect.ChainHandler) {
	s.chain = chain
}

func (s *StorageService) Name() string {
	return "storage"
}

func (s *StorageService) DownloadFile(ctx context.Context, req *connect.Request[v1.DownloadFileRequest]) (*connect.Response[v1.DownloadFileResponse], error) {
	data, err := s.localStore.GetTranscoded(req.Msg.Cid)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("file not found: %w", err))
	}

	return connect.NewResponse(&v1.DownloadFileResponse{
		Data: data,
	}), nil
}

func (s *StorageService) DownloadFileChunk(ctx context.Context, req *connect.Request[v1.DownloadFileChunkRequest], stream *connect.ServerStream[v1.DownloadFileChunkResponse]) error {
	data, err := s.localStore.GetTranscoded(req.Msg.Cid)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("file not found: %w", err))
	}

	chunkSize := int(req.Msg.ChunkSize)
	if chunkSize <= 0 {
		chunkSize = MaxChunkSize
	}

	var chunkIndex uint32
	for offset := 0; offset < len(data); offset += chunkSize {
		end := offset + chunkSize
		if end > len(data) {
			end = len(data)
		}

		if err := stream.Send(&v1.DownloadFileChunkResponse{
			Data:       data[offset:end],
			ChunkIndex: chunkIndex,
			IsLast:     end == len(data),
		}); err != nil {
			return err
		}
		chunkIndex++
	}

	return nil
}

func (s *StorageService) Upload(ctx context.Context, req *connect.Request[v1.UploadRequest]) (*connect.Response[v1.UploadResponse], error) {
	data := req.Msg.Data
	expectedCID := req.Msg.Cid
	meta := req.Msg.Metadata

	// Validate size
	if len(data) > MaxDirectUploadSize {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("file too large: %d bytes (max %d)", len(data), MaxDirectUploadSize))
	}

	// Validate CID
	if err := cid.Validate(expectedCID, data); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("CID validation failed: %w", err))
	}

	// Store upload with metadata
	uploadMeta := &storagev1.UploadMeta{
		FileName: meta.FileName,
		MimeType: meta.MimeType,
		Size:     meta.Size,
	}
	if err := s.localStore.StoreUpload(expectedCID, data, uploadMeta); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to store upload: %w", err))
	}

	// Transcode
	transcodedCID, err := s.transcodeFile(ctx, expectedCID, data, meta.MimeType)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("transcoding failed: %w", err))
	}

	// Submit transaction
	if err := s.submitFileUploadTx(ctx, expectedCID, transcodedCID, meta); err != nil {
		s.Logger.Warnf("failed to submit file upload tx: %v", err)
		// Don't fail the upload, tx can be retried
	}

	return connect.NewResponse(&v1.UploadResponse{
		OriginalCid:   expectedCID,
		TranscodedCid: transcodedCID,
	}), nil
}

func (s *StorageService) UploadChunk(ctx context.Context, req *connect.Request[v1.UploadChunkRequest]) (*connect.Response[v1.UploadChunkResponse], error) {
	data := req.Msg.Data
	expectedCID := req.Msg.Cid
	chunkIndex := req.Msg.ChunkIndex
	totalChunks := req.Msg.TotalChunks
	meta := req.Msg.Metadata

	// Validate chunk size
	if len(data) > MaxChunkSize {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("chunk too large: %d bytes (max %d)", len(data), MaxChunkSize))
	}

	// Build upload meta (only first chunk or when metadata is provided)
	var uploadMeta *storagev1.UploadMeta
	if meta != nil {
		uploadMeta = &storagev1.UploadMeta{
			FileName:    meta.FileName,
			MimeType:    meta.MimeType,
			Size:        meta.Size,
			TotalChunks: totalChunks,
		}
	} else if chunkIndex == 0 {
		// First chunk must have metadata
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("first chunk must include metadata"))
	}

	// Store chunk
	if err := s.localStore.StoreChunk(expectedCID, chunkIndex, data, uploadMeta); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to store chunk: %w", err))
	}

	// Check if all chunks are uploaded
	hasAll, err := s.localStore.HasAllChunks(expectedCID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to check chunks: %w", err))
	}

	if !hasAll {
		// Return progress
		chunksReceived, err := s.localStore.CountChunks(expectedCID)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to count chunks: %w", err))
		}
		storedMeta, _ := s.localStore.GetUploadMeta(expectedCID)
		var total uint32
		if storedMeta != nil {
			total = storedMeta.TotalChunks
		}
		return connect.NewResponse(&v1.UploadChunkResponse{
			Complete:       false,
			ChunksReceived: chunksReceived,
			TotalChunks:    total,
		}), nil
	}

	// All chunks received - reassemble
	fullData, err := s.localStore.ReassembleChunks(expectedCID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to reassemble chunks: %w", err))
	}

	// Validate CID of reassembled file
	if err := cid.Validate(expectedCID, fullData); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("reassembled file CID validation failed: %w", err))
	}

	// Get stored metadata
	storedMeta, err := s.localStore.GetUploadMeta(expectedCID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get metadata: %w", err))
	}

	// Store the complete file
	if err := s.localStore.StoreUpload(expectedCID, fullData, storedMeta); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to store reassembled file: %w", err))
	}

	// Delete chunks
	if err := s.localStore.DeleteChunks(expectedCID); err != nil {
		s.Logger.Warnf("failed to delete chunks: %v", err)
	}

	// Transcode
	transcodedCID, err := s.transcodeFile(ctx, expectedCID, fullData, storedMeta.MimeType)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("transcoding failed: %w", err))
	}

	// Submit transaction
	fileMeta := &v1.FileMetadata{
		FileName: storedMeta.FileName,
		MimeType: storedMeta.MimeType,
		Size:     storedMeta.Size,
	}
	if err := s.submitFileUploadTx(ctx, expectedCID, transcodedCID, fileMeta); err != nil {
		s.Logger.Warnf("failed to submit file upload tx: %v", err)
	}

	return connect.NewResponse(&v1.UploadChunkResponse{
		Complete:       true,
		ChunksReceived: storedMeta.TotalChunks,
		TotalChunks:    storedMeta.TotalChunks,
		OriginalCid:    expectedCID,
		TranscodedCid:  transcodedCID,
	}), nil
}

func (s *StorageService) transcodeFile(ctx context.Context, originalCID string, data []byte, mimeType string) (string, error) {
	var out bytes.Buffer
	in := bytes.NewReader(data)

	if strings.HasPrefix(mimeType, "audio/") {
		if err := s.encoder.EncodeAudio(ctx, in, &out); err != nil {
			return "", fmt.Errorf("audio encoding failed: %w", err)
		}
	} else if strings.HasPrefix(mimeType, "image/") {
		if err := s.encoder.EncodeImage(ctx, in, &out); err != nil {
			return "", fmt.Errorf("image encoding failed: %w", err)
		}
	} else {
		return "", fmt.Errorf("unsupported media type: %s", mimeType)
	}

	transcodedData := out.Bytes()
	transcodedCID, err := cid.Compute(transcodedData)
	if err != nil {
		return "", fmt.Errorf("failed to compute transcoded CID: %w", err)
	}

	if err := s.localStore.StoreTranscoded(transcodedCID, transcodedData); err != nil {
		return "", fmt.Errorf("failed to store transcoded file: %w", err)
	}

	return transcodedCID, nil
}

func (s *StorageService) submitFileUploadTx(ctx context.Context, originalCID, transcodedCID string, meta *v1.FileMetadata) error {
	if s.chain == nil {
		return fmt.Errorf("chain client not set")
	}

	// Build the transaction
	uploaderAddr := "" // TODO: Get from request context/auth
	transcoderAddr := s.config.Mojave.ValidatorAddress

	msg := &storagev1.FileUploadMessage{
		UploaderAddress:   uploaderAddr,
		TranscoderAddress: transcoderAddr,
		OriginalCid:       originalCID,
		TranscodedCid:     transcodedCID,
		FileName:          meta.FileName,
		MimeType:          meta.MimeType,
		Size:              meta.Size,
	}

	signedTx := &chainv1.SignedTransaction{
		Transaction: &chainv1.Transaction{
			Header: &chainv1.TransactionHeader{
				ChainId:   s.config.Mojave.ChainID,
				Nonce:     uint64(time.Now().UnixNano()),
				GasPrice:  1,
				GasLimit:  100000,
				Timeout:   uint64(time.Now().Add(time.Hour).Unix()),
				Sender:    transcoderAddr,
				Recipient: "",
			},
			Body: &chainv1.TransactionBody{
				Body: &chainv1.TransactionBody_FileUpload{
					FileUpload: &chainv1.FileUploadTransaction{
						Msg: msg,
					},
				},
			},
		},
		Signature: &chainv1.TransactionSignature{
			Signature: []byte("validator-signature"), // TODO: Proper signing
		},
	}

	txBytes, err := proto.Marshal(signedTx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	req := connect.NewRequest(&v1.SendTransactionRequest{
		SignedTransaction: txBytes,
	})

	_, err = s.chain.SendTransaction(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	return nil
}

var _ v1connect.StorageHandler = (*StorageService)(nil)

func NewStorageService(
	config *config.Config,
	logger *zap.Logger,
	localStore *localstore.LocalStore,
	chainStore *chainstore.ChainStore,
) (*StorageService, error) {
	encoder, err := media.NewMediaEncoder(MaxEncoderWorkers)
	if err != nil {
		return nil, fmt.Errorf("failed to create media encoder: %w", err)
	}

	svc := &StorageService{
		config:     config,
		localStore: localStore,
		chainStore: chainStore,
		encoder:    encoder,
	}
	svc.BaseModule = module.NewBaseModule(logger.Named(svc.Name()))
	return svc, nil
}

// ABCI++ Callbacks

func (s *StorageService) CheckTx(ctx context.Context, req *abcitypes.CheckTxRequest) (*abcitypes.CheckTxResponse, error) {
	return &abcitypes.CheckTxResponse{}, nil
}

func (s *StorageService) PrepareProposal(ctx context.Context, req *abcitypes.PrepareProposalRequest) (*abcitypes.PrepareProposalResponse, error) {
	return &abcitypes.PrepareProposalResponse{}, nil
}

func (s *StorageService) ProcessProposal(ctx context.Context, req *abcitypes.ProcessProposalRequest) (*abcitypes.ProcessProposalResponse, error) {
	return &abcitypes.ProcessProposalResponse{}, nil
}

func (s *StorageService) FinalizeBlock(ctx context.Context, req *abcitypes.FinalizeBlockRequest) (*abcitypes.FinalizeBlockResponse, error) {
	for _, txBytes := range req.Txs {
		var signedTx chainv1.SignedTransaction
		if err := proto.Unmarshal(txBytes, &signedTx); err != nil {
			continue
		}

		if signedTx.Transaction == nil || signedTx.Transaction.Body == nil {
			continue
		}

		if fileUpload := signedTx.Transaction.Body.GetFileUpload(); fileUpload != nil {
			msg := fileUpload.Msg
			if msg == nil {
				continue
			}

			// Store in chainstore
			if err := s.chainStore.StoreUpload(msg); err != nil {
				s.Logger.Errorf("failed to store upload in chainstore: %v", err)
				continue
			}

			// Delete original upload from localstore
			if err := s.localStore.DeleteUpload(msg.OriginalCid); err != nil {
				s.Logger.Warnf("failed to delete original upload: %v", err)
			}

			s.Logger.Infof("finalized file upload: original=%s transcoded=%s", msg.OriginalCid, msg.TranscodedCid)
		}
	}

	return &abcitypes.FinalizeBlockResponse{}, nil
}
