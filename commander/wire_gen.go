// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/qx66/camp/internal/biz"
	"github.com/qx66/camp/internal/conf"
	"github.com/qx66/camp/internal/data"
	"github.com/qx66/camp/internal/service"
	"go.uber.org/zap"
)

// Injectors from wire.go:

func initApp(logger *zap.Logger, data2 *conf.Data) (*app, func(), error) {
	dataData, cleanup, err := data.NewData(data2, logger)
	if err != nil {
		return nil, nil, err
	}
	instructRepo := data.NewInstructDataSource(dataData)
	instanceRepo := data.NewInstanceDataSource(dataData)
	messageUseCase := biz.NewMessageUseCase(logger, instructRepo, instanceRepo)
	instructUseCase := biz.NewInstructUseCase(instructRepo, logger)
	instanceUseCase := biz.NewInstanceUseCase(instanceRepo, logger)
	useCase := service.NewUseCase(logger, messageUseCase, instructUseCase, instanceUseCase)
	mainApp := newApp(useCase)
	return mainApp, func() {
		cleanup()
	}, nil
}
