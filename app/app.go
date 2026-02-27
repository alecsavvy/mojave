package app

import (
	"context"
	"log"
	"os"
	"path"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/proxy"
	"github.com/dgraph-io/badger/v4"
	"go.uber.org/zap"

	"github.com/alecsavvy/mojave/config"
	cmtflags "github.com/cometbft/cometbft/libs/cli/flags"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	nm "github.com/cometbft/cometbft/node"
)

type App struct {
	logger *zap.SugaredLogger
	node   *nm.Node
}

func NewApp(homeDir string) *App {
	z, _ := zap.NewDevelopment()
	logger := z.Sugar()

	cmtConfig := cfg.DefaultConfig()
	cmtConfig.SetRoot(homeDir)

	pv, nodeKey, _, err := config.InitFilesWithConfig(cmtConfig)
	if err != nil {
		panic(err)
	}

	if err := cmtConfig.ValidateBasic(); err != nil {
		panic(err)
	}

	dbPath := path.Join(homeDir, "badger")

	db, err := badger.Open(badger.DefaultOptions(dbPath))
	if err != nil {
		panic(err)
	}

	cmtLogger := cmtlog.NewTMLogger(cmtlog.NewSyncWriter(os.Stdout))
	cmtLogger, err = cmtflags.ParseLogLevel(cmtConfig.LogLevel, cmtLogger, cfg.DefaultLogLevel)
	if err != nil {
		panic(err)
	}

	abci := NewKVStoreApplication(logger, db)

	node, err := nm.NewNode(
		context.Background(),
		cmtConfig,
		pv,
		nodeKey,
		proxy.NewLocalClientCreator(abci),
		nm.DefaultGenesisDocProviderFunc(cmtConfig),
		cfg.DefaultDBProvider,
		nm.DefaultMetricsProvider(cmtConfig.Instrumentation),
		cmtLogger,
	)

	if err != nil {
		log.Fatalf("Creating node: %v", err)
	}

	if err := node.Start(); err != nil {
		log.Fatalf("Starting node: %v", err)
	}

	return &App{
		logger: logger,
		node:   node,
	}
}

func (a *App) Run(ctx context.Context) error {
	go func() { <-ctx.Done(); _ = a.node.Stop() }()
	a.node.Wait()
	return nil
}
