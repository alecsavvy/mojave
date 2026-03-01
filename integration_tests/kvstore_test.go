package integrationtests

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/alecsavvy/mojave/sdk"
	"github.com/stretchr/testify/require"
)

func TestKVStore(t *testing.T) {
	ctx := t.Context()

	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	sdk, err := sdk.NewMojaveSDK("http://localhost:26657")
	if err != nil {
		t.Fatalf("Failed to create SDK: %v", err)
	}
	sdk.SetPrivateKey(privKey)

	err = sdk.SetKeyValue(ctx, "cometbft", "rocks")
	if err != nil {
		t.Fatalf("Failed to set key-value: %v", err)
	}

	kvState, err := sdk.GetKeyValue(ctx, "cometbft")
	if err != nil {
		t.Fatalf("Failed to get key-value: %v", err)
	}

	require.Equal(t, "cometbft", kvState.Key)
	require.Equal(t, "rocks", kvState.Value)
}
