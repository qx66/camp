package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/qx66/camp/internal/conf"
	"github.com/qx66/camp/internal/service"
	"go.uber.org/zap"
)

type app struct {
	service *service.UseCase
}

func newApp(service *service.UseCase) *app {
	return &app{
		service: service,
	}
}

var (
	configPath = ""
)

func init() {
	flag.StringVar(&configPath, "configPath", "./configs/config.yaml", "")
}

func main() {
	flag.Parse()
	
	logger, err := zap.NewProduction()
	if err != nil {
		return
	}
	
	var configs []config.Source
	// 本地配置文件配置
	logger.Info("加载本地配置文件", zap.String("configPath", configPath))
	configs = append(configs, file.NewSource(configPath))
	
	c := config.New(
		config.WithSource(
			configs...,
			//file.NewSource(flagconf),
		),
	)
	
	if err := c.Load(); err != nil {
		panic(err)
	}
	
	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}
	
	app, clean, err := initApp(logger, bc.Data)
	defer clean()
	
	if err != nil {
		logger.Error("初始化程序失败", zap.Error(err))
		return
	}
	
	g := gin.New()
	
	g.GET("connect", app.service.Connect)
	g.GET("instance/alive", app.service.ListAliveInstance)
	//
	g.POST("instruct", app.service.Instruct)
	g.GET("instruct", app.service.ListInstruct)
	g.GET("instruct/:uuid", app.service.GetInstruct)
	
	err = g.Run(bc.Server.Http.Addr)
	
	logger.Error("程序异常", zap.Error(err))
}
