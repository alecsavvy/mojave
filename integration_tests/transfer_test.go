package integrationtests

import (
	"math"
	"testing"

	"github.com/alecsavvy/mojave/utils"
	"github.com/stretchr/testify/require"
)

func TestTransfer(t *testing.T) {
	ctx := t.Context()

	app := StartTestApp(ctx, t.TempDir())
	t.Cleanup(func() {
		app.Stop()
	})

	sdk := app.SDK()

	account, err := sdk.GetAccount(ctx, utils.ZeroAddress)
	if err != nil {
		t.Fatalf("failed to get account: %v", err)
	}
	require.Equal(t, uint64(math.MaxUint64), account.Balance)

	err = sdk.FaucetTokens(ctx, sdk.GetPublicKey(), 1000)
	if err != nil {
		t.Fatalf("failed to faucet tokens: %v", err)
	}

	account, err = sdk.GetAccount(ctx, sdk.GetPublicKey())
	if err != nil {
		t.Fatalf("failed to get account: %v", err)
	}
	require.Equal(t, uint64(1000), account.Balance)

	sdk2 := app.SDK()

	if err := sdk.TransferTokens(ctx, sdk.GetPublicKey(), sdk2.GetPublicKey(), 100); err != nil {
		t.Fatalf("failed to transfer tokens: %v", err)
	}

	account, err = sdk.GetAccount(ctx, sdk.GetPublicKey())
	if err != nil {
		t.Fatalf("failed to get account: %v", err)
	}
	require.Equal(t, uint64(900), account.Balance)

	account, err = sdk2.GetAccount(ctx, sdk2.GetPublicKey())
	if err != nil {
		t.Fatalf("failed to get account: %v", err)
	}
	require.Equal(t, uint64(100), account.Balance)
}
