package app

import (
	"context"
	"log"
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
	"golang.org/x/sync/errgroup"
)

type App struct {
	config *config.Config
	node   *cmtnm.Node
	core   *core.Core

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
func NewApp(config *config.Config) (*App, error) {
	chain := chain.NewChainService(config)
	storage := storage.NewStorageService(config)
	system := system.NewSystemService(config)
	p2p := p2p.NewP2PService(config)
	ddex := ddex.NewDDEXService(config)
	composition := composition.NewCompositionService(config)
	account := account.NewAccountService(config)
	validator := validator.NewValidatorService(config)

	chainStore, err := pebble_chainstore.NewPebbleChainStore(config.Sonata.ChainStore.Path)
	if err != nil {
		return nil, err
	}

	localStore, err := pebble_localstore.NewPebbleLocalStore(config.Sonata.LocalStore.Path)
	if err != nil {
		return nil, err
	}

	cmtConfig := config.CometBFT

	pv := privval.LoadFilePV(
		cmtConfig.PrivValidatorKeyFile(),
		cmtConfig.PrivValidatorStateFile(),
	)

	nodeKey, err := cmtp2p.LoadNodeKey(cmtConfig.NodeKeyFile())
	if err != nil {
		log.Fatalf("failed to load node's key: %v", err)
	}

	logger := cmtlog.NewTMLogger(cmtlog.NewSyncWriter(os.Stdout))
	logger, err = cmtflags.ParseLogLevel(cmtConfig.LogLevel, logger, cmtconfig.DefaultLogLevel)
	if err != nil {
		return nil, err
	}

	createNode := func(c *core.Core) (*cmtnm.Node, error) {
		return cmtnm.NewNode(context.Background(), cmtConfig, pv, nodeKey, proxy.NewLocalClientCreator(c),
			cmtnm.DefaultGenesisDocProviderFunc(cmtConfig),
			cmtconfig.DefaultDBProvider,
			cmtnm.DefaultMetricsProvider(cmtConfig.Instrumentation), logger)
	}

	core, node, err := core.NewCore(config, createNode, chain, storage, system, p2p, ddex, composition, account, validator)
	if err != nil {
		return nil, err
	}

	server, err := server.NewServer(config, chain, storage, system, p2p, ddex, composition, account, validator)
	if err != nil {
		return nil, err
	}

	// TODO: wire up dependencies

	return &App{
		core:   core,
		config: config,
		node:   node,

		chain:       chain,
		server:      server,
		storage:     storage,
		system:      system,
		p2p:         p2p,
		ddex:        ddex,
		composition: composition,
		account:     account,
		validator:   validator,

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
		log.Printf("shutting down...\n")
		return app.Shutdown()
	})

	return eg.Wait()
}

func (app *App) Shutdown() error {

	eg, _ := errgroup.WithContext(context.Background())
	defer func() {
		log.Printf("shutdown complete\n")
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
