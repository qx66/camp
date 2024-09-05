//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/qx66/camp/internal/biz"
	"github.com/qx66/camp/internal/conf"
	"github.com/qx66/camp/internal/data"
	"github.com/qx66/camp/internal/service"
	"go.uber.org/zap"
)

func initApp(logger *zap.Logger, data2 *conf.Data) (*app, func(), error) {
	panic(wire.Build(data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
