package app

import (
	"context"
	"log"
	"path"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	"github.com/dgraph-io/badger/v4"
	"go.uber.org/zap"

	cmtlog "github.com/cometbft/cometbft/libs/log"
	nm "github.com/cometbft/cometbft/node"
)

type App struct {
	logger *zap.SugaredLogger
	node   *nm.Node
}

// NewApp starts a node from an already-initialized config. The caller must have
// written config.toml, genesis.json, priv validator key/state, and node key to the config's RootDir.
func NewApp(cmtConfig *cfg.Config) *App {
	z, _ := zap.NewDevelopment()
	logger := z.Sugar()

	if err := cmtConfig.ValidateBasic(); err != nil {
		panic(err)
	}

	pv := privval.LoadFilePV(cmtConfig.PrivValidatorKeyFile(), cmtConfig.PrivValidatorStateFile())
	nodeKey, err := p2p.LoadNodeKey(cmtConfig.NodeKeyFile())
	if err != nil {
		log.Fatalf("Load node key: %v", err)
	}

	dbPath := path.Join(cmtConfig.RootDir, "badger")
	db, err := badger.Open(badger.DefaultOptions(dbPath))
	if err != nil {
		panic(err)
	}

	cmtLogger := cmtlog.NewNopLogger()

	addr := pv.GetAddress().String()
	logger = logger.With("addr", addr)

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
