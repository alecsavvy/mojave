package app

import (
	"context"
	"path"

	"github.com/alecsavvy/mojave/store"
	"github.com/cockroachdb/pebble"
	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
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
func NewApp(cmtConfig *cfg.Config) (*App, error) {
	z, _ := zap.NewDevelopment()
	logger := z.Sugar()

	if err := cmtConfig.ValidateBasic(); err != nil {
		return nil, err
	}

	pv := privval.LoadFilePV(cmtConfig.PrivValidatorKeyFile(), cmtConfig.PrivValidatorStateFile())
	nodeKey, err := p2p.LoadNodeKey(cmtConfig.NodeKeyFile())
	if err != nil {
		return nil, err
	}

	dbPath := path.Join(cmtConfig.RootDir, "pebble")
	db, err := pebble.Open(dbPath, nil)
	if err != nil {
		return nil, err
	}

	cmtLogger := cmtlog.NewNopLogger()

	addr := pv.GetAddress().String()
	logger = logger.With("addr", addr)

	appStore := store.NewStore(db)
	abci := NewKVStoreApplication(logger, appStore)

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
		return nil, err
	}
	if err := node.Start(); err != nil {
		return nil, err
	}

	return &App{
		logger: logger,
		node:   node,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	go func() { <-ctx.Done(); _ = a.node.Stop() }()
	a.node.Wait()
	return nil
}
