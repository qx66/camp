//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/qx66/camp/internal/biz"
	"go.uber.org/zap"
)

func initApp(logger *zap.Logger) *app {
	panic(wire.Build(biz.ProviderSet, newApp))
}
