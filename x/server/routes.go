package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
)

func (s *Server) registerRoutes() {
	httpServer := s.httpServer

	httpServer.HideBanner = true

	httpServer.Use(middleware.Logger(), middleware.Recover())

	httpServer.GET("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]uint{"a": 440})
	})

	rpcGroup := httpServer.Group("")
	chainPath, chainHandler := v1connect.NewChainHandler(s.chain)
	rpcGroup.Any(chainPath, echo.WrapHandler(chainHandler))

	storagePath, storageHandler := v1connect.NewStorageHandler(s.storage)
	rpcGroup.Any(storagePath, echo.WrapHandler(storageHandler))

	systemPath, systemHandler := v1connect.NewSystemHandler(s.system)
	rpcGroup.Any(systemPath, echo.WrapHandler(systemHandler))

	p2pPath, p2pHandler := v1connect.NewP2PHandler(s.p2p)
	rpcGroup.Any(p2pPath, echo.WrapHandler(p2pHandler))

	ddexPath, ddexHandler := v1connect.NewDDEXHandler(s.ddex)
	rpcGroup.Any(ddexPath, echo.WrapHandler(ddexHandler))

	compositionPath, compositionHandler := v1connect.NewCompositionHandler(s.composition)
	rpcGroup.Any(compositionPath, echo.WrapHandler(compositionHandler))

	accountPath, accountHandler := v1connect.NewAccountHandler(s.account)
	rpcGroup.Any(accountPath, echo.WrapHandler(accountHandler))

	validatorPath, validatorHandler := v1connect.NewValidatorHandler(s.validator)
	rpcGroup.Any(validatorPath, echo.WrapHandler(validatorHandler))
}
