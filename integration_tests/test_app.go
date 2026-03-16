package integrationtests

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alecsavvy/mojave/app"
	"github.com/alecsavvy/mojave/config"
	"github.com/alecsavvy/mojave/sdk"
	cfg "github.com/cometbft/cometbft/config"
)

type TestApp struct {
	connectAddr string
	app         *app.App
}

func StartTestApp(ctx context.Context, homeDir string) *TestApp {
	cmtConfig := cfg.DefaultConfig()
	cmtConfig.SetRoot(homeDir)

	if _, _, _, err := config.InitFilesWithConfig(cmtConfig); err != nil {
		panic(err)
	}

	// Blob buckets: upload writes final TDF to storageBucket; temp file goes to RootDir/upload-tmp.
	dataTmp := filepath.Join(homeDir, "data", "tmp")
	dataStorage := filepath.Join(homeDir, "data", "audio")
	if err := os.MkdirAll(dataTmp, 0755); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(dataStorage, 0755); err != nil {
		panic(err)
	}
	absTmp, _ := filepath.Abs(dataTmp)
	absStorage, _ := filepath.Abs(dataStorage)

	mojaveConfig := &config.MojaveConfig{
		CometConfig:     cmtConfig,
		ConnectRPCAddr:  "127.0.0.1:9090",
		FilesTmpDir:     "file://" + absTmp,
		FilesStorageDir: "file://" + absStorage,
	}
	a, err := app.NewApp(mojaveConfig)
	if err != nil {
		panic(err)
	}

	if err := a.Start(); err != nil {
		panic(err)
	}

	testApp := &TestApp{
		connectAddr: "http://127.0.0.1:9090",
		app:         a,
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
	s, err := sdk.NewMojaveSDK(node.connectAddr)
	if err != nil {
		panic(err)
	}

	s.SetPrivateKey(privKey)
	return s
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
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
		latest, err := node.app.LatestBlockHeight(ctx)
		if err != nil {
			continue
		}
		if latest >= height {
			return nil
		}
	}
}
