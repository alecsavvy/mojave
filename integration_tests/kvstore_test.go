package integrationtests

import (
	"testing"

	"github.com/alecsavvy/mojave/sdk"
	"github.com/stretchr/testify/require"
)

func TestKVStore(t *testing.T) {
	ctx := t.Context()

	sdk, err := sdk.NewMojaveSDK("http://localhost:26657")
	if err != nil {
		t.Fatalf("Failed to create SDK: %v", err)
	}

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
