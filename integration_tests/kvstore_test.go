package integrationtests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKVStore(t *testing.T) {
	ctx := t.Context()

	app := StartTestApp(ctx, t.TempDir())
	sdk := app.SDK()

	err := sdk.SetKeyValue(ctx, "cometbft", "rocks")
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
