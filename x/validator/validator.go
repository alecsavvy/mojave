package validator

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
	"github.com/sonata-labs/sonata/types/module"
	"go.uber.org/zap"
)

type ValidatorService struct {
	*module.BaseModule
	config *config.Config
}

func (v *ValidatorService) Name() string {
	return "validator"
}

func (v *ValidatorService) GetValidator(context.Context, *connect.Request[v1.GetValidatorRequest]) (*connect.Response[v1.GetValidatorResponse], error) {
	panic("unimplemented")
}

func (v *ValidatorService) GetValidators(context.Context, *connect.Request[v1.GetValidatorsRequest]) (*connect.Response[v1.GetValidatorsResponse], error) {
	panic("unimplemented")
}

var _ v1connect.ValidatorHandler = (*ValidatorService)(nil)

func NewValidatorService(config *config.Config, logger *zap.Logger) *ValidatorService {
	svc := &ValidatorService{config: config}
	svc.BaseModule = module.NewBaseModule(logger.Named(svc.Name()))
	return svc
}
