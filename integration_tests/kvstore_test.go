package integrationtests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKVStore(t *testing.T) {
	ctx := t.Context()

	app := StartTestApp(ctx, t.TempDir())
	t.Cleanup(func() {
		app.Stop()
	})
	sdk := app.SDK()

	_, err := sdk.SetKeyValue(ctx, "cometbft", "rocks")
	require.NoError(t, err)

	kvState, err := sdk.GetKeyValue(ctx, "cometbft")
	require.NoError(t, err)

	require.Equal(t, "cometbft", kvState.Key)
	require.Equal(t, "rocks", kvState.Value)
}
