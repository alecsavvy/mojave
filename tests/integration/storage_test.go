package integration

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/common/cid"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/sdk"
)

// TestDirectUpload tests the direct upload flow for files < 50MB.
func TestDirectUpload(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	nodeURL := getNodeURL()
	client := sdk.NewSonataSDK(nodeURL)

	// Generate test data (small audio-like bytes)
	testData := make([]byte, 1024*100) // 100KB test file
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	// Compute CID client-side
	expectedCID, err := cid.Compute(testData)
	if err != nil {
		t.Fatalf("failed to compute CID: %v", err)
	}
	t.Logf("computed CID: %s", expectedCID)

	// Upload the file
	uploadReq := connect.NewRequest(&v1.UploadRequest{
		Cid:  expectedCID,
		Data: testData,
		Metadata: &v1.FileMetadata{
			FileName: "test-audio.flac",
			MimeType: "audio/flac",
			Size:     uint64(len(testData)),
		},
	})

	uploadResp, err := client.Storage.Upload(ctx, uploadReq)
	if err != nil {
		t.Fatalf("failed to upload file: %v", err)
	}

	// Verify response
	if uploadResp.Msg.OriginalCid != expectedCID {
		t.Errorf("original CID mismatch: got %s, want %s", uploadResp.Msg.OriginalCid, expectedCID)
	}

	if uploadResp.Msg.TranscodedCid == "" {
		t.Error("transcoded CID should not be empty")
	}

	t.Logf("upload successful: original=%s, transcoded=%s", uploadResp.Msg.OriginalCid, uploadResp.Msg.TranscodedCid)

	// Wait for block finalization
	time.Sleep(2 * time.Second)

	// Verify we can download the transcoded file
	downloadReq := connect.NewRequest(&v1.DownloadFileRequest{
		Cid: uploadResp.Msg.TranscodedCid,
	})

	downloadResp, err := client.Storage.DownloadFile(ctx, downloadReq)
	if err != nil {
		t.Fatalf("failed to download file: %v", err)
	}

	if len(downloadResp.Msg.Data) == 0 {
		t.Error("downloaded file should not be empty")
	}

	t.Logf("download successful: %d bytes", len(downloadResp.Msg.Data))
}

// TestChunkedUpload tests the chunked upload flow for larger files.
func TestChunkedUpload(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	nodeURL := getNodeURL()
	client := sdk.NewSonataSDK(nodeURL)

	// Generate test data (larger file that would be chunked)
	testData := make([]byte, 1024*1024*5) // 5MB test file
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	// Compute CID of full file client-side
	expectedCID, err := cid.Compute(testData)
	if err != nil {
		t.Fatalf("failed to compute CID: %v", err)
	}
	t.Logf("computed full file CID: %s", expectedCID)

	// Split into chunks (1MB each for testing)
	chunkSize := 1024 * 1024 // 1MB
	var chunks [][]byte
	for i := 0; i < len(testData); i += chunkSize {
		end := i + chunkSize
		if end > len(testData) {
			end = len(testData)
		}
		chunks = append(chunks, testData[i:end])
	}

	totalChunks := uint32(len(chunks))
	t.Logf("split into %d chunks", totalChunks)

	// Upload chunks (test out-of-order by uploading last chunk first)
	uploadOrder := []int{len(chunks) - 1} // Start with last chunk
	for i := 0; i < len(chunks)-1; i++ {
		uploadOrder = append(uploadOrder, i)
	}

	var finalResp *connect.Response[v1.UploadChunkResponse]
	for _, idx := range uploadOrder {
		chunk := chunks[idx]
		chunkIndex := uint32(idx)

		req := &v1.UploadChunkRequest{
			Cid:         expectedCID,
			ChunkIndex:  chunkIndex,
			TotalChunks: totalChunks,
			Data:        chunk,
		}

		// Only include metadata with first chunk (index 0)
		if idx == 0 {
			req.Metadata = &v1.FileMetadata{
				FileName: "test-large-audio.flac",
				MimeType: "audio/flac",
				Size:     uint64(len(testData)),
			}
		}

		resp, err := client.Storage.UploadChunk(ctx, connect.NewRequest(req))
		if err != nil {
			t.Fatalf("failed to upload chunk %d: %v", idx, err)
		}

		t.Logf("uploaded chunk %d/%d, complete=%v, received=%d",
			chunkIndex+1, totalChunks, resp.Msg.Complete, resp.Msg.ChunksReceived)

		if resp.Msg.Complete {
			finalResp = resp
		}
	}

	// Verify final response
	if finalResp == nil {
		t.Fatal("upload should be complete after all chunks")
	}

	if !finalResp.Msg.Complete {
		t.Error("final response should indicate completion")
	}

	if finalResp.Msg.OriginalCid != expectedCID {
		t.Errorf("original CID mismatch: got %s, want %s", finalResp.Msg.OriginalCid, expectedCID)
	}

	if finalResp.Msg.TranscodedCid == "" {
		t.Error("transcoded CID should not be empty")
	}

	t.Logf("chunked upload successful: original=%s, transcoded=%s",
		finalResp.Msg.OriginalCid, finalResp.Msg.TranscodedCid)

	// Wait for block finalization
	time.Sleep(2 * time.Second)

	// Verify we can download the transcoded file
	downloadReq := connect.NewRequest(&v1.DownloadFileRequest{
		Cid: finalResp.Msg.TranscodedCid,
	})

	downloadResp, err := client.Storage.DownloadFile(ctx, downloadReq)
	if err != nil {
		t.Fatalf("failed to download file: %v", err)
	}

	if len(downloadResp.Msg.Data) == 0 {
		t.Error("downloaded file should not be empty")
	}

	t.Logf("download successful: %d bytes", len(downloadResp.Msg.Data))
}

// TestChunkedUploadInOrder tests chunked upload with in-order delivery.
func TestChunkedUploadInOrder(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	nodeURL := getNodeURL()
	client := sdk.NewSonataSDK(nodeURL)

	// Generate test data
	testData := make([]byte, 1024*1024*3) // 3MB test file
	for i := range testData {
		testData[i] = byte((i * 7) % 256) // Different pattern
	}

	// Compute CID of full file
	expectedCID, err := cid.Compute(testData)
	if err != nil {
		t.Fatalf("failed to compute CID: %v", err)
	}

	// Split into chunks
	chunkSize := 1024 * 1024 // 1MB
	var chunks [][]byte
	for i := 0; i < len(testData); i += chunkSize {
		end := i + chunkSize
		if end > len(testData) {
			end = len(testData)
		}
		chunks = append(chunks, testData[i:end])
	}

	totalChunks := uint32(len(chunks))

	// Upload chunks in order
	var finalResp *connect.Response[v1.UploadChunkResponse]
	for idx, chunk := range chunks {
		req := &v1.UploadChunkRequest{
			Cid:         expectedCID,
			ChunkIndex:  uint32(idx),
			TotalChunks: totalChunks,
			Data:        chunk,
		}

		if idx == 0 {
			req.Metadata = &v1.FileMetadata{
				FileName: "test-ordered-audio.flac",
				MimeType: "audio/flac",
				Size:     uint64(len(testData)),
			}
		}

		resp, err := client.Storage.UploadChunk(ctx, connect.NewRequest(req))
		if err != nil {
			t.Fatalf("failed to upload chunk %d: %v", idx, err)
		}

		if resp.Msg.Complete {
			finalResp = resp
		}
	}

	if finalResp == nil || !finalResp.Msg.Complete {
		t.Fatal("upload should be complete")
	}

	if finalResp.Msg.TranscodedCid == "" {
		t.Error("transcoded CID should not be empty")
	}

	t.Logf("in-order chunked upload successful: transcoded=%s", finalResp.Msg.TranscodedCid)
}

