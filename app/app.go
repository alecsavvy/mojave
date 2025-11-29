package app

import (
	"context"
	"os"

	cmtconfig "github.com/cometbft/cometbft/config"
	cmtflags "github.com/cometbft/cometbft/libs/cli/flags"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	cmtnm "github.com/cometbft/cometbft/node"
	cmtp2p "github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	"github.com/sonata-labs/sonata/config"
	"github.com/sonata-labs/sonata/core"
	"github.com/sonata-labs/sonata/store/chainstore"
	pebble_chainstore "github.com/sonata-labs/sonata/store/chainstore/pebble"
	"github.com/sonata-labs/sonata/store/localstore"
	pebble_localstore "github.com/sonata-labs/sonata/store/localstore/pebble"
	"github.com/sonata-labs/sonata/x/account"
	"github.com/sonata-labs/sonata/x/chain"
	"github.com/sonata-labs/sonata/x/composition"
	"github.com/sonata-labs/sonata/x/ddex"
	"github.com/sonata-labs/sonata/x/p2p"
	"github.com/sonata-labs/sonata/x/server"
	"github.com/sonata-labs/sonata/x/storage"
	"github.com/sonata-labs/sonata/x/system"
	"github.com/sonata-labs/sonata/x/validator"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type App struct {
	config *config.Config
	node   *cmtnm.Node
	core   *core.Core
	logger *zap.SugaredLogger

	server *server.Server

	chain       *chain.ChainService
	storage     *storage.StorageService
	system      *system.SystemService
	p2p         *p2p.P2PService
	ddex        *ddex.DDEXService
	composition *composition.CompositionService
	account     *account.AccountService
	validator   *validator.ValidatorService

	chainStore chainstore.ChainStore
	localStore localstore.LocalStore
}

// Creates and initializes all modules and the app
func NewApp(cfg *config.Config, zapLogger *zap.Logger) (*App, error) {
	appLogger := zapLogger.Named("app").Sugar()

	chainSvc := chain.NewChainService(cfg, zapLogger)
	storageSvc := storage.NewStorageService(cfg, zapLogger)
	systemSvc := system.NewSystemService(cfg, zapLogger)
	p2pSvc := p2p.NewP2PService(cfg, zapLogger)
	ddexSvc := ddex.NewDDEXService(cfg, zapLogger)
	compositionSvc := composition.NewCompositionService(cfg, zapLogger)
	accountSvc := account.NewAccountService(cfg, zapLogger)
	validatorSvc := validator.NewValidatorService(cfg, zapLogger)

	chainStore, err := pebble_chainstore.NewPebbleChainStore(cfg.Sonata.ChainStore.Path)
	if err != nil {
		return nil, err
	}

	localStore, err := pebble_localstore.NewPebbleLocalStore(cfg.Sonata.LocalStore.Path)
	if err != nil {
		return nil, err
	}

	cmtConfig := cfg.CometBFT

	pv := privval.LoadFilePV(
		cmtConfig.PrivValidatorKeyFile(),
		cmtConfig.PrivValidatorStateFile(),
	)

	nodeKey, err := cmtp2p.LoadNodeKey(cmtConfig.NodeKeyFile())
	if err != nil {
		appLogger.Fatalf("failed to load node's key: %v", err)
	}

	cmtLogger := cmtlog.NewTMLogger(cmtlog.NewSyncWriter(os.Stdout))
	cmtLogger, err = cmtflags.ParseLogLevel(cmtConfig.LogLevel, cmtLogger, cmtconfig.DefaultLogLevel)
	if err != nil {
		return nil, err
	}

	createNode := func(c *core.Core) (*cmtnm.Node, error) {
		return cmtnm.NewNode(context.Background(), cmtConfig, pv, nodeKey, proxy.NewLocalClientCreator(c),
			cmtnm.DefaultGenesisDocProviderFunc(cmtConfig),
			cmtconfig.DefaultDBProvider,
			cmtnm.DefaultMetricsProvider(cmtConfig.Instrumentation), cmtLogger)
	}

	coreSvc, node, err := core.NewCore(cfg, zapLogger, createNode, chainSvc, storageSvc, systemSvc, p2pSvc, ddexSvc, compositionSvc, accountSvc, validatorSvc)
	if err != nil {
		return nil, err
	}

	serverSvc, err := server.NewServer(cfg, zapLogger, chainSvc, storageSvc, systemSvc, p2pSvc, ddexSvc, compositionSvc, accountSvc, validatorSvc)
	if err != nil {
		return nil, err
	}

	return &App{
		core:   coreSvc,
		config: cfg,
		node:   node,
		logger: appLogger,

		chain:       chainSvc,
		server:      serverSvc,
		storage:     storageSvc,
		system:      systemSvc,
		p2p:         p2pSvc,
		ddex:        ddexSvc,
		composition: compositionSvc,
		account:     accountSvc,
		validator:   validatorSvc,

		chainStore: chainStore,
		localStore: localStore,
	}, nil
}

func (app *App) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	// start up all modules
	eg.Go(app.server.Start)
	eg.Go(app.core.Start)
	eg.Go(app.ddex.Start)
	eg.Go(app.composition.Start)
	eg.Go(app.account.Start)
	eg.Go(app.validator.Start)
	eg.Go(app.chain.Start)
	eg.Go(app.storage.Start)
	eg.Go(app.system.Start)
	eg.Go(app.p2p.Start)

	eg.Go(func() error {
		<-ctx.Done()
		app.logger.Info("shutting down...")
		return app.Shutdown()
	})

	return eg.Wait()
}

func (app *App) Shutdown() error {

	eg, _ := errgroup.WithContext(context.Background())
	defer func() {
		app.logger.Info("shutdown complete")
	}()

	// shutdown all modules
	eg.Go(app.server.Stop)
	eg.Go(app.core.Stop)
	eg.Go(app.ddex.Stop)
	eg.Go(app.composition.Stop)
	eg.Go(app.account.Stop)
	eg.Go(app.validator.Stop)
	eg.Go(app.chain.Stop)
	eg.Go(app.storage.Stop)
	eg.Go(app.system.Stop)
	eg.Go(app.p2p.Stop)

	return eg.Wait()
}
