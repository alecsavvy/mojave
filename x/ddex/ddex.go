package ddex

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
	"github.com/sonata-labs/sonata/types/module"
	"go.uber.org/zap"
)

type DDEXService struct {
	*module.BaseModule
	config *config.Config
}

func (d *DDEXService) Name() string {
	return "ddex"
}

var _ module.Module = (*DDEXService)(nil)

func (d *DDEXService) GetCatalogListMessage(context.Context, *connect.Request[v1.GetCatalogListMessageRequest]) (*connect.Response[v1.GetCatalogListMessageResponse], error) {
	panic("unimplemented")
}

func (d *DDEXService) GetMeadMessage(context.Context, *connect.Request[v1.GetMeadMessageRequest]) (*connect.Response[v1.GetMeadMessageResponse], error) {
	panic("unimplemented")
}

func (d *DDEXService) GetNewReleaseMessage(context.Context, *connect.Request[v1.GetNewReleaseMessageRequest]) (*connect.Response[v1.GetNewReleaseMessageResponse], error) {
	panic("unimplemented")
}

func (d *DDEXService) GetPieMessage(context.Context, *connect.Request[v1.GetPieMessageRequest]) (*connect.Response[v1.GetPieMessageResponse], error) {
	panic("unimplemented")
}

func (d *DDEXService) GetPieRequestMessage(context.Context, *connect.Request[v1.GetPieRequestMessageRequest]) (*connect.Response[v1.GetPieRequestMessageResponse], error) {
	panic("unimplemented")
}

func (d *DDEXService) GetPurgeReleaseMessage(context.Context, *connect.Request[v1.GetPurgeReleaseMessageRequest]) (*connect.Response[v1.GetPurgeReleaseMessageResponse], error) {
	panic("unimplemented")
}

var _ v1connect.DDEXHandler = (*DDEXService)(nil)

func NewDDEXService(config *config.Config, logger *zap.Logger) *DDEXService {
	svc := &DDEXService{config: config}
	svc.BaseModule = module.NewBaseModule(logger.Named(svc.Name()))
	return svc
}
