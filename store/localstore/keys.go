package localstore

import "fmt"

const (
	UploadPrefix     = "upload/"
	UploadMetaPrefix = "upload_meta/"
	ChunkPrefix      = "chunk/"
	TranscodedPrefix = "transcoded/"
)

func uploadKey(cid string) []byte {
	return []byte(UploadPrefix + cid)
}

func uploadMetaKey(cid string) []byte {
	return []byte(UploadMetaPrefix + cid)
}

func chunkKey(cid string, index uint32) []byte {
	return []byte(fmt.Sprintf("%s%s/%s%d", UploadPrefix, cid, ChunkPrefix, index))
}

func transcodedKey(cid string) []byte {
	return []byte(TranscodedPrefix + cid)
}

