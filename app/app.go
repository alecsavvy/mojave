package app

import (
	"context"

	"go.uber.org/zap"
)

type App struct {
	logger *zap.SugaredLogger
}

func NewApp() *App {
	logger, _ := zap.NewDevelopment()
	return &App{
		logger: logger.Sugar(),
	}
}

func (a *App) Run(ctx context.Context) error {
	a.logger.Info("good morning!")
	defer a.logger.Info("goodnight!")
	return nil
}
