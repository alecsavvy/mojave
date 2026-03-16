package integrationtests

import (
	"crypto/sha256"
	"os/exec"
	"testing"

	v1 "github.com/alecsavvy/mojave/gen/mojave/v1"
	"github.com/stretchr/testify/require"
)

// minimalWAV returns a minimal valid PCM WAV file (mono, 8kHz, 8-bit) that ffmpeg can transcode.
func minimalWAV() []byte {
	// RIFF header: "RIFF" + size (52) + "WAVE"
	// fmt subchunk: "fmt " + 16 + PCM(1) + mono(1) + 8000Hz + 8000 byte/sec + 1 block + 8 bits
	// data subchunk: "data" + size(4) + 4 bytes PCM
	const (
		riffSize = 52 // 60 - 8
		dataLen  = 4
	)
	header := []byte{
		'R', 'I', 'F', 'F',
		riffSize & 0xff, riffSize >> 8, 0, 0,
		'W', 'A', 'V', 'E',
		'f', 'm', 't', ' ',
		16, 0, 0, 0, // fmt size
		1, 0, // PCM
		1, 0, // mono
		0x40, 0x1f, 0, 0, // 8000 Hz
		0x40, 0x1f, 0, 0, // byte rate
		1, 0, // block align
		8, 0, // bits per sample
		'd', 'a', 't', 'a',
		dataLen & 0xff, dataLen >> 8, 0, 0,
	}
	pcm := []byte{0x80, 0x80, 0x80, 0x80} // silence
	return append(header, pcm...)
}

// TestUploadFile runs an end-to-end upload: transcode to FLAC, encrypt to TDF, broadcast FileUploadTransaction.
// Requires ffmpeg in PATH; skips if not found so CI can run in environments without ffmpeg.
func TestUploadFile(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not in PATH, skipping upload integration test")
	}

	ctx := t.Context()
	dir := t.TempDir()

	app := StartTestApp(ctx, dir)
	t.Cleanup(func() {
		app.Stop()
	})

	s := app.SDK()
	fileData := minimalWAV()
	hash := sha256.Sum256(fileData)

	req := &v1.UploadFileRequest{
		UploaderPubkey: s.GetPublicKey(),
		FileHash:       hash[:],
		FileData:       fileData,
	}

	result, err := s.UploadFile(ctx, req)
	require.NoError(t, err)
	require.NotEmpty(t, result.Infohash)
	require.Greater(t, result.FileSize, uint64(0))

	// GetFile by infohash returns record and deterministic magnet URI
	fileResp, err := s.GetFile(ctx, result.Infohash)
	require.NoError(t, err)
	require.Equal(t, result.Infohash, fileResp.File.Infohash)
	require.Equal(t, "magnet:?xt=urn:btih:"+result.Infohash, fileResp.MagnetUri)
}
