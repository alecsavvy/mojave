package app

import (
	"context"
	"os"

	"github.com/alecsavvy/mojave/config"
	"github.com/alecsavvy/mojave/core"
	"github.com/alecsavvy/mojave/store/chainstore"
	"github.com/alecsavvy/mojave/store/localstore"
	"github.com/alecsavvy/mojave/x/account"
	"github.com/alecsavvy/mojave/x/chain"
	"github.com/alecsavvy/mojave/x/composition"
	"github.com/alecsavvy/mojave/x/ddex"
	"github.com/alecsavvy/mojave/x/p2p"
	"github.com/alecsavvy/mojave/x/server"
	"github.com/alecsavvy/mojave/x/statesync"
	"github.com/alecsavvy/mojave/x/storage"
	"github.com/alecsavvy/mojave/x/system"
	"github.com/alecsavvy/mojave/x/validator"
	cmtconfig "github.com/cometbft/cometbft/config"
	cmtflags "github.com/cometbft/cometbft/libs/cli/flags"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	cmtnm "github.com/cometbft/cometbft/node"
	cmtp2p "github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
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
	statesync   *statesync.StateSyncService

	chainStore *chainstore.ChainStore
	localStore *localstore.LocalStore
}

// Creates and initializes all modules and the app
func NewApp(cfg *config.Config, zapLogger *zap.Logger) (*App, error) {
	appLogger := zapLogger.Named("app").Sugar()

	chainStore, err := chainstore.NewChainStore(cfg.Mojave.ChainStore.Path)
	if err != nil {
		return nil, err
	}

	localStore, err := localstore.NewLocalStore(cfg.Mojave.LocalStore.Path)
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

	coreSvc, node, err := core.NewCore(cfg, zapLogger, createNode, chainStore)
	if err != nil {
		return nil, err
	}

	chainSvc := chain.NewChainService(cfg, zapLogger, node)
	storageSvc, err := storage.NewStorageService(cfg, zapLogger, localStore, chainStore)
	if err != nil {
		return nil, err
	}

	// Set dependencies
	storageSvc.SetChain(chainSvc)
	systemSvc := system.NewSystemService(cfg, zapLogger)
	p2pSvc := p2p.NewP2PService(cfg, zapLogger)
	ddexSvc := ddex.NewDDEXService(cfg, zapLogger)
	compositionSvc := composition.NewCompositionService(cfg, zapLogger)
	accountSvc := account.NewAccountService(cfg, zapLogger, chainStore)
	validatorSvc := validator.NewValidatorService(cfg, zapLogger)
	statesyncSvc := statesync.NewStateSyncService(cfg, zapLogger, chainStore)

	coreSvc.RegisterModules(core.InitChain, chainSvc)
	coreSvc.RegisterModules(core.CheckTx, chainSvc, accountSvc, ddexSvc, storageSvc, compositionSvc, validatorSvc)
	coreSvc.RegisterModules(core.PrepareProposal, chainSvc, storageSvc, systemSvc, ddexSvc, compositionSvc, accountSvc, validatorSvc)
	coreSvc.RegisterModules(core.ProcessProposal, chainSvc, storageSvc, systemSvc, ddexSvc, compositionSvc, accountSvc, validatorSvc)
	coreSvc.RegisterModules(core.FinalizeBlock, chainSvc, storageSvc, systemSvc, ddexSvc, compositionSvc, accountSvc, validatorSvc, statesyncSvc)
	coreSvc.RegisterModules(core.Commit, chainSvc, statesyncSvc)

	coreSvc.RegisterModules(core.ListSnapshots, statesyncSvc)
	coreSvc.RegisterModules(core.OfferSnapshot, statesyncSvc)
	coreSvc.RegisterModules(core.LoadSnapshotChunk, statesyncSvc)
	coreSvc.RegisterModules(core.ApplySnapshotChunk, statesyncSvc)

	serverSvc, err := server.NewServer(cfg, zapLogger, chainSvc, storageSvc, systemSvc, p2pSvc, ddexSvc, compositionSvc, accountSvc, validatorSvc)
	if err != nil {
		return nil, err
	}

	// Register startup dependencies
	// Core waits for all modules to be ready before starting CometBFT
	coreSvc.RegisterStartupDeps(
		chainSvc.Ready(), storageSvc.Ready(), systemSvc.Ready(),
		p2pSvc.Ready(), ddexSvc.Ready(), compositionSvc.Ready(),
		accountSvc.Ready(), validatorSvc.Ready(), statesyncSvc.Ready(),
	)
	// Server waits for core to be ready before accepting HTTP traffic
	serverSvc.RegisterStartupDeps(coreSvc.Ready())

	// Register shutdown dependencies (reverse order of startup)
	// Server shuts down first (no deps)
	// Core waits for server to stop before shutting down
	coreSvc.RegisterShutdownDeps(serverSvc.Stopped())

	// All modules wait for core to stop before shutting down
	chainSvc.RegisterShutdownDeps(coreSvc.Stopped())
	storageSvc.RegisterShutdownDeps(coreSvc.Stopped())
	systemSvc.RegisterShutdownDeps(coreSvc.Stopped())
	p2pSvc.RegisterShutdownDeps(coreSvc.Stopped())
	ddexSvc.RegisterShutdownDeps(coreSvc.Stopped())
	compositionSvc.RegisterShutdownDeps(coreSvc.Stopped())
	accountSvc.RegisterShutdownDeps(coreSvc.Stopped())
	validatorSvc.RegisterShutdownDeps(coreSvc.Stopped())
	statesyncSvc.RegisterShutdownDeps(coreSvc.Stopped())

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
		statesync:   statesyncSvc,

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
	eg.Go(app.statesync.Start)

	eg.Go(func() error {
		<-ctx.Done()
		app.logger.Info("shutting down...")
		return app.Shutdown()
	})

	return eg.Wait()
}

func (app *App) Shutdown() error {
	eg, _ := errgroup.WithContext(context.Background())
	defer app.logger.Info("shutdown complete")

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
	eg.Go(app.statesync.Stop)

	return eg.Wait()
}
