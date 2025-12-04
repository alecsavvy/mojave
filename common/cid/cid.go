package cid

import (
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

// Compute generates a CIDv1 from data using SHA-256 multihash.
func Compute(data []byte) (string, error) {
	hash := sha256.Sum256(data)
	mh, err := multihash.Encode(hash[:], multihash.SHA2_256)
	if err != nil {
		return "", fmt.Errorf("failed to encode multihash: %w", err)
	}

	c := cid.NewCidV1(cid.Raw, mh)
	return c.String(), nil
}

// ComputeFromReader generates a CIDv1 from a reader using SHA-256 multihash.
func ComputeFromReader(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("failed to read data: %w", err)
	}

	mh, err := multihash.Encode(h.Sum(nil), multihash.SHA2_256)
	if err != nil {
		return "", fmt.Errorf("failed to encode multihash: %w", err)
	}

	c := cid.NewCidV1(cid.Raw, mh)
	return c.String(), nil
}

// Validate verifies that data matches the expected CID.
func Validate(cidStr string, data []byte) error {
	expected, err := cid.Decode(cidStr)
	if err != nil {
		return fmt.Errorf("invalid CID: %w", err)
	}

	hash := sha256.Sum256(data)
	mh, err := multihash.Encode(hash[:], multihash.SHA2_256)
	if err != nil {
		return fmt.Errorf("failed to encode multihash: %w", err)
	}

	actual := cid.NewCidV1(cid.Raw, mh)
	if !expected.Equals(actual) {
		return fmt.Errorf("CID mismatch: expected %s, got %s", expected.String(), actual.String())
	}

	return nil
}

// Parse decodes a CID string and returns the parsed CID.
func Parse(cidStr string) (cid.Cid, error) {
	return cid.Decode(cidStr)
}

