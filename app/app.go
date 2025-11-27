package app

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sonata-labs/sonata/account"
	"github.com/sonata-labs/sonata/composition"
	"github.com/sonata-labs/sonata/config"
	"github.com/sonata-labs/sonata/core"
	"github.com/sonata-labs/sonata/ddex"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
	"github.com/sonata-labs/sonata/p2p"
	"github.com/sonata-labs/sonata/storage"
	"github.com/sonata-labs/sonata/store/appstore"
	"github.com/sonata-labs/sonata/store/localstore"
	"github.com/sonata-labs/sonata/system"
	"github.com/sonata-labs/sonata/validator"
	"golang.org/x/sync/errgroup"
)

type App struct {
	config *config.Config

	httpServer *echo.Echo

	core        v1connect.CoreHandler
	storage     v1connect.StorageHandler
	system      v1connect.SystemHandler
	p2p         v1connect.P2PHandler
	ddex        v1connect.DDEXHandler
	composition v1connect.CompositionHandler
	account     v1connect.AccountHandler
	validator   v1connect.ValidatorHandler

	appstore   v1connect.AppStoreHandler
	localstore v1connect.LocalStoreHandler
}

func NewApp(config *config.Config) *App {
	core := core.NewCoreService(config)
	storage := storage.NewStorageService(config)
	system := system.NewSystemService(config)
	p2p := p2p.NewP2PService(config)
	ddex := ddex.NewDDEXService(config)
	composition := composition.NewCompositionService(config)
	account := account.NewAccountService(config)
	validator := validator.NewValidatorService(config)

	appstore := appstore.NewAppStoreService(config)
	localstore := localstore.NewLocalStoreService(config)

	httpServer := echo.New()

	httpServer.HideBanner = true

	httpServer.Use(middleware.Logger(), middleware.Recover())

	httpServer.GET("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]uint{"a": 440})
	})

	rpcGroup := httpServer.Group("")
	corePath, coreHandler := v1connect.NewCoreHandler(core)
	rpcGroup.Any(corePath, echo.WrapHandler(coreHandler))

	storagePath, storageHandler := v1connect.NewStorageHandler(storage)
	rpcGroup.Any(storagePath, echo.WrapHandler(storageHandler))

	systemPath, systemHandler := v1connect.NewSystemHandler(system)
	rpcGroup.Any(systemPath, echo.WrapHandler(systemHandler))

	p2pPath, p2pHandler := v1connect.NewP2PHandler(p2p)
	rpcGroup.Any(p2pPath, echo.WrapHandler(p2pHandler))

	ddexPath, ddexHandler := v1connect.NewDDEXHandler(ddex)
	rpcGroup.Any(ddexPath, echo.WrapHandler(ddexHandler))

	compositionPath, compositionHandler := v1connect.NewCompositionHandler(composition)
	rpcGroup.Any(compositionPath, echo.WrapHandler(compositionHandler))

	accountPath, accountHandler := v1connect.NewAccountHandler(account)
	rpcGroup.Any(accountPath, echo.WrapHandler(accountHandler))

	validatorPath, validatorHandler := v1connect.NewValidatorHandler(validator)
	rpcGroup.Any(validatorPath, echo.WrapHandler(validatorHandler))

	return &App{
		config:      config,
		httpServer:  httpServer,
		core:        core,
		storage:     storage,
		system:      system,
		p2p:         p2p,
		ddex:        ddex,
		composition: composition,
		account:     account,
		validator:   validator,
		appstore:    appstore,
		localstore:  localstore,
	}
}

func (app *App) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		if err := app.runHTTP(ctx); err != nil {
			return err
		}
		return nil
	})

	eg.Go(func() error {
		<-ctx.Done()
		log.Printf("shutting down...")
		return app.Shutdown()
	})

	return eg.Wait()
}

func (app *App) Shutdown() error {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	wg := sync.WaitGroup{}

	wg.Go(func() {
		if err := app.httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("failed to shutdown HTTP server: %v", err)
		}
	})

	wg.Wait()

	return nil
}
