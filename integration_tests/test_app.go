package integrationtests

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/alecsavvy/mojave/app"
	"github.com/alecsavvy/mojave/config"
	"github.com/alecsavvy/mojave/sdk"
	cfg "github.com/cometbft/cometbft/config"
)

type TestApp struct {
	config *cfg.Config
	app    *app.App
}

func StartTestApp(ctx context.Context, homeDir string) *TestApp {
	cmtConfig := cfg.DefaultConfig()
	cmtConfig.SetRoot(homeDir)

	if _, _, _, err := config.InitFilesWithConfig(cmtConfig); err != nil {
		panic(err)
	}

	a, err := app.NewApp(cmtConfig)
	if err != nil {
		panic(err)
	}

	if err := a.Start(); err != nil {
		panic(err)
	}

	testApp := &TestApp{
		config: cmtConfig,
		app:    a,
	}

	if err := testApp.AwaitBlockHeight(ctx, 1); err != nil {
		panic(err)
	}

	return testApp
}

// SDK returns a new SDK for the test app with a random private key
func (node *TestApp) SDK() *sdk.MojaveSDK {
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	sdk, err := sdk.NewMojaveSDK(node.config.RPC.ListenAddress)
	if err != nil {
		panic(err)
	}

	sdk.SetPrivateKey(privKey)
	return sdk
}

func (node *TestApp) Start() error {
	if err := node.app.Start(); err != nil {
		return fmt.Errorf("failed to run app: %w", err)
	}
	return nil
}

func (node *TestApp) Stop() error {
	return node.app.Stop()
}

func (node *TestApp) AwaitBlockHeight(ctx context.Context, height int64) error {
	sdk := node.SDK()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
		res, err := sdk.Status(ctx)
		if err != nil {
			return err
		}
		if res.SyncInfo.LatestBlockHeight >= height {
			return nil
		}
	}
}
