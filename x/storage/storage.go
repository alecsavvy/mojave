package storage

import (
	"context"

	"connectrpc.com/connect"
	"github.com/sonata-labs/sonata/config"
	v1 "github.com/sonata-labs/sonata/gen/api/v1"
	"github.com/sonata-labs/sonata/gen/api/v1/v1connect"
	"github.com/sonata-labs/sonata/types/module"
	"go.uber.org/zap"
)

type StorageService struct {
	*module.BaseModule
	config *config.Config
}

func (s *StorageService) Name() string {
	return "storage"
}

func (s *StorageService) DownloadFile(context.Context, *connect.Request[v1.DownloadFileRequest]) (*connect.Response[v1.DownloadFileResponse], error) {
	panic("unimplemented")
}

func (s *StorageService) DownloadFileChunk(context.Context, *connect.Request[v1.DownloadFileChunkRequest], *connect.ServerStream[v1.DownloadFileChunkResponse]) error {
	panic("unimplemented")
}

func (s *StorageService) Upload(context.Context, *connect.Request[v1.UploadRequest]) (*connect.Response[v1.UploadResponse], error) {
	panic("unimplemented")
}

func (s *StorageService) UploadChunk(context.Context, *connect.Request[v1.UploadChunkRequest]) (*connect.Response[v1.UploadChunkResponse], error) {
	panic("unimplemented")
}

var _ v1connect.StorageHandler = (*StorageService)(nil)

func NewStorageService(config *config.Config, logger *zap.Logger) *StorageService {
	svc := &StorageService{config: config}
	svc.BaseModule = module.NewBaseModule(logger.Named(svc.Name()))
	return svc
}
