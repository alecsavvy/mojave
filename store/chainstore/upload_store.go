package chainstore

import (
	storagev1 "github.com/alecsavvy/mojave/gen/storage/v1"
	"google.golang.org/protobuf/proto"
)

const (
	UploadPrefix         = "upload/"
	UploadOriginalPrefix = "upload_idx/original/"
)

func uploadKey(cid string) []byte {
	return []byte(UploadPrefix + cid)
}

func uploadOriginalIndexKey(originalCID string) []byte {
	return []byte(UploadOriginalPrefix + originalCID)
}

// StoreUpload stores a file upload record and creates an index by original CID.
func (c *ChainStore) StoreUpload(upload *storagev1.FileUploadMessage) error {
	if err := c.RequireBatch(); err != nil {
		return err
	}

	uploadBytes, err := proto.Marshal(upload)
	if err != nil {
		return err
	}

	// Store by transcoded CID (primary key)
	if err := c.writer.Set(uploadKey(upload.TranscodedCid), uploadBytes, nil); err != nil {
		return err
	}

	// Create index by original CID pointing to transcoded CID
	if err := c.writer.Set(uploadOriginalIndexKey(upload.OriginalCid), []byte(upload.TranscodedCid), nil); err != nil {
		return err
	}

	return nil
}

// GetUpload retrieves a file upload record by transcoded CID.
func (c *ChainStore) GetUpload(cid string) (*storagev1.FileUploadMessage, error) {
	data, closer, err := c.reader.Get(uploadKey(cid))
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	upload := &storagev1.FileUploadMessage{}
	if err := proto.Unmarshal(data, upload); err != nil {
		return nil, err
	}

	return upload, nil
}

// GetUploadByOriginalCID retrieves a file upload record by original CID.
func (c *ChainStore) GetUploadByOriginalCID(originalCID string) (*storagev1.FileUploadMessage, error) {
	// Look up the transcoded CID from the index
	transcodedCIDBytes, closer, err := c.reader.Get(uploadOriginalIndexKey(originalCID))
	if err != nil {
		return nil, err
	}
	closer.Close()

	// Get the actual upload record
	return c.GetUpload(string(transcodedCIDBytes))
}
