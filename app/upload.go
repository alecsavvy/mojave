package app

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"connectrpc.com/connect"
	mcrypto "github.com/alecsavvy/mojave/crypto"
	v1 "github.com/alecsavvy/mojave/gen/mojave/v1"
	"github.com/alecsavvy/opentdf/pkg/opentdf"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

func (app *App) CheckFileUploadTx(ctx context.Context, tx *v1.FileUploadTransaction) error {
	return nil
}

func (app *App) FinalizeFileUploadTx(ctx context.Context, tx *v1.FileUploadTransaction) error {
	return nil
}

// UploadFile handles an upload request by transcoding, encrypting, storing, seeding, and recording it on-chain.
func (app *App) UploadFile(ctx context.Context, req *connect.Request[v1.UploadFileRequest]) (*connect.Response[v1.UploadFileResponse], error) {
	uploadTmpDir := filepath.Join(app.node.Config().RootDir, "upload-tmp")
	if err := os.MkdirAll(uploadTmpDir, 0755); err != nil {
		return nil, err
	}

	baseName := hex.EncodeToString(req.Msg.FileHash)
	tdfPath := filepath.Join(uploadTmpDir, baseName+".tdf")

	tdfFileSize, err := app.transcodeAndEncryptToTDF(ctx, req.Msg.FileData, tdfPath)
	if err != nil {
		return nil, err
	}

	infohash, err := app.buildInfohashFromTDF(tdfPath)
	if err != nil {
		return nil, err
	}

	if err := app.addAndSeedTorrent(ctx, infohash, tdfPath); err != nil {
		return nil, err
	}

	storageKey, err := app.storeTDF(ctx, tdfPath, infohash)
	if err != nil {
		return nil, err
	}

	if err := app.buildAndSendUploadTx(ctx, req.Msg, infohash, tdfFileSize); err != nil {
		_ = app.storageBucket.Delete(ctx, storageKey)
		return nil, err
	}

	return connect.NewResponse(&v1.UploadFileResponse{
		FileUploadResult: &v1.FileUploadResult{
			Infohash: infohash,
			FileSize: uint64(tdfFileSize),
		},
	}), nil
}

// transcodeAndEncryptToTDF runs ffmpeg to transcode the input to FLAC and encrypts it into a TDF file at tdfPath.
func (app *App) transcodeAndEncryptToTDF(ctx context.Context, data []byte, tdfPath string) (int64, error) {
	tdfFile, err := os.Create(tdfPath)
	if err != nil {
		return 0, err
	}
	defer tdfFile.Close()

	cmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-i", "pipe:0",
		"-f", "flac", "-c:a", "flac", "-compression_level", "9", "pipe:1")
	cmd.Stdin = bytes.NewReader(data)
	flacOut, err := cmd.StdoutPipe()
	if err != nil {
		return 0, fmt.Errorf("ffmpeg stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("ffmpeg start: %w", err)
	}

	err = opentdf.EncryptTo(tdfFile, flacOut, opentdf.EncryptConfig{
		Locator:            app.node.GenesisDoc().ChainID,
		AuthorityPublicKey: &app.encryptionKey.PublicKey,
		MIMEType:           "audio/flac",
	})
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return 0, fmt.Errorf("opentdf encrypt: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return 0, fmt.Errorf("ffmpeg: %w", err)
	}

	tdfStat, err := os.Stat(tdfPath)
	if err != nil {
		return 0, err
	}

	return tdfStat.Size(), nil
}

// buildInfohashFromTDF computes the torrent infohash for the TDF file at tdfPath.
func (app *App) buildInfohashFromTDF(tdfPath string) (string, error) {
	var info metainfo.Info
	info.PieceLength = 256 * 1024
	if err := info.BuildFromFilePath(tdfPath); err != nil {
		return "", fmt.Errorf("torrent info: %w", err)
	}
	infoBytes, err := bencode.Marshal(&info)
	if err != nil {
		return "", err
	}
	mi := metainfo.MetaInfo{InfoBytes: infoBytes}
	return mi.HashInfoBytes().HexString(), nil
}

// addAndSeedTorrent registers the TDF with the local torrent client and starts seeding it.
func (app *App) addAndSeedTorrent(ctx context.Context, infohash, tdfPath string) error {
	if app.torrentClient == nil {
		return nil
	}

	// TODO: In the future, make seeding behavior configurable for non-validators.
	var info metainfo.Info
	info.PieceLength = 256 * 1024
	if err := info.BuildFromFilePath(tdfPath); err != nil {
		return fmt.Errorf("torrent info: %w", err)
	}
	infoBytes, err := bencode.Marshal(&info)
	if err != nil {
		return err
	}
	mi := metainfo.MetaInfo{InfoBytes: infoBytes}

	if _, err := app.torrentClient.AddTorrent(&mi); err != nil {
		return fmt.Errorf("torrent add: %w", err)
	}

	return nil
}

// storeTDF writes the TDF file at tdfPath to the storage bucket and returns the storage key.
func (app *App) storeTDF(ctx context.Context, tdfPath, infohash string) (string, error) {
	storageKey := infohash + ".tdf"

	f, err := os.Open(tdfPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	w, err := app.storageBucket.NewWriter(ctx, storageKey, nil)
	if err != nil {
		return "", fmt.Errorf("storage bucket writer: %w", err)
	}

	if _, err := io.Copy(w, f); err != nil {
		_ = w.Close()
		_ = app.storageBucket.Delete(ctx, storageKey)
		return "", fmt.Errorf("copy to storage: %w", err)
	}

	if err := w.Close(); err != nil {
		_ = app.storageBucket.Delete(ctx, storageKey)
		return "", fmt.Errorf("close storage writer: %w", err)
	}

	return storageKey, nil
}

// buildAndSendUploadTx constructs, signs, and sends a file upload transaction for the given infohash and file size.
func (app *App) buildAndSendUploadTx(ctx context.Context, req *v1.UploadFileRequest, infohash string, tdfFileSize int64) error {
	tx := &v1.Transaction{
		Header: &v1.TransactionHeader{
			ChainId:    app.node.GenesisDoc().ChainID,
			Nonce:      hex.EncodeToString(req.FileHash),
			FromPubkey: app.validatorPubKey,
		},
		Body: &v1.TransactionBody{
			Body: &v1.TransactionBody_FileUpload{
				FileUpload: &v1.FileUploadTransaction{
					UploaderPubkey: req.UploaderPubkey,
					Infohash:       infohash,
					FileSize:       uint64(tdfFileSize),
				},
			},
		},
	}

	signed, err := mcrypto.SignTransaction(app.validatorPrivKey, tx)
	if err != nil {
		return fmt.Errorf("sign tx: %w", err)
	}

	_, err = app.SendTransaction(ctx, connect.NewRequest(&v1.SendTransactionRequest{SignedTransaction: signed}))
	if err != nil {
		return fmt.Errorf("send tx: %w", err)
	}

	return nil
}
