package app

import (
	"context"
	"net"
	"net/http"
	"path"

	"github.com/alecsavvy/mojave/config"
	"github.com/alecsavvy/mojave/gen/mojave/v1/v1connect"
	"github.com/alecsavvy/mojave/store"
	"github.com/cockroachdb/pebble"
	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	"github.com/cometbft/cometbft/rpc/client/local"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	cmtlog "github.com/cometbft/cometbft/libs/log"
	nm "github.com/cometbft/cometbft/node"
)

type App struct {
	logger *zap.SugaredLogger
	node   *nm.Node
	rpc    *local.Local

	store         *store.Store
	onGoingBlock  *pebble.Batch
	connectServer *http.Server
}

// NewApp starts a node from an already-initialized config. The caller must have
// written config.toml, genesis.json, priv validator key/state, and node key to the config's RootDir.
func NewApp(mojaveCfg *config.MojaveConfig) (*App, error) {
	cmtConfig := mojaveCfg.CometConfig

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
	app := &App{
		logger:       logger,
		store:        appStore,
		onGoingBlock: nil,
	}

	node, err := nm.NewNode(
		context.Background(),
		cmtConfig,
		pv,
		nodeKey,
		proxy.NewLocalClientCreator(app),
		nm.DefaultGenesisDocProviderFunc(cmtConfig),
		cfg.DefaultDBProvider,
		nm.DefaultMetricsProvider(cmtConfig.Instrumentation),
		cmtLogger,
	)
	if err != nil {
		return nil, err
	}

	app.node = node
	app.rpc = local.New(app.node)

	mux := http.NewServeMux()
	svcPath, svcHandler := v1connect.NewServiceHandler(app)
	mux.Handle(svcPath, svcHandler)
	app.connectServer = &http.Server{
		Addr:    mojaveCfg.ConnectRPCAddr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	return app, nil
}

func (a *App) Start() error {
	if a.node.IsRunning() {
		return nil
	}
	if err := a.node.Start(); err != nil {
		return err
	}

	ln, err := net.Listen("tcp", a.connectServer.Addr)
	if err != nil {
		return err
	}
	a.logger.Infow("ConnectRPC server listening", "addr", a.connectServer.Addr)
	go func() {
		if err := a.connectServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			a.logger.Errorw("ConnectRPC server error", "err", err)
		}
	}()

	return nil
}

func (a *App) LatestBlockHeight(ctx context.Context) (int64, error) {
	status, err := a.rpc.Status(ctx)
	if err != nil {
		return 0, err
	}
	return status.SyncInfo.LatestBlockHeight, nil
}

func (a *App) Stop() error {
	if err := a.connectServer.Shutdown(context.Background()); err != nil {
		a.logger.Warnw("ConnectRPC server shutdown error", "err", err)
	}
	return a.node.Stop()
}
