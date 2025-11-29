package logger

import (
	"go.uber.org/zap"

	cmtlog "github.com/cometbft/cometbft/libs/log"
)

type CometBFTAdapter struct {
	logger *zap.SugaredLogger
}

func NewCometBFTAdapter(logger *zap.Logger) cmtlog.Logger {
	return &CometBFTAdapter{logger: logger.Named("cometbft").Sugar()}
}

func (z *CometBFTAdapter) Debug(msg string, keyvals ...interface{}) {
	z.logger.Debugw(msg, keyvals...)
}

func (z *CometBFTAdapter) Info(msg string, keyvals ...interface{}) {
	z.logger.Infow(msg, keyvals...)
}

func (z *CometBFTAdapter) Error(msg string, keyvals ...interface{}) {
	z.logger.Errorw(msg, keyvals...)
}

func (z *CometBFTAdapter) With(keyvals ...interface{}) cmtlog.Logger {
	return &CometBFTAdapter{logger: z.logger.With(keyvals...)}
}

