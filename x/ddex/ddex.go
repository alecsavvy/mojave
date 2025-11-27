package ddex

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
)

type DDEXService struct {
	config *config.Config
}

var _ v1connect.DDEXHandler = (*DDEXService)(nil)

// Ping implements v1connect.DDEXHandler.
func (d *DDEXService) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	panic("unimplemented")
}

func NewDDEXService(config *config.Config) *DDEXService {
	return &DDEXService{
		config: config,
	}
}
