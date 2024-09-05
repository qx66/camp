package data

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"github.com/qx66/camp/internal/conf"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewInstructDataSource, NewInstanceDataSource)

// Data .
type Data struct {
	db     *gorm.DB
	redis  *redis.Client
	logger *zap.Logger
}

// NewData .
func NewData(c *conf.Data, logger *zap.Logger) (*Data, func(), error) {
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})
	if err != nil {
		logger.Error("failed opening connection to mysql", zap.Error(err))
		return nil, nil, err
	}
	//db.Use()
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("return db err", zap.Error(err))
		return nil, nil, err
	}
	
	err = sqlDB.Ping()
	if err != nil {
		logger.Error("ping dataSource failure", zap.Error(err))
		return nil, nil, err
	}
	
	sqlDB.SetMaxIdleConns(int(c.Database.MaxIdleConns))
	sqlDB.SetMaxOpenConns(int(c.Database.MaxOpenConns))
	
	// redis
	redis := redis.NewClient(&redis.Options{
		DB:           int(c.Redis.Db),
		Network:      c.Redis.Network,
		Addr:         c.Redis.Addr,
		Password:     c.Redis.Password,
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
		PoolSize:     int(c.Redis.PoolSize),
		MinIdleConns: int(c.Redis.MinIdleConns),
		DialTimeout:  c.Redis.DialTimeout.AsDuration(),
	})
	
	err = redis.Ping(context.Background()).Err()
	if err != nil {
		logger.Error("ping dataSource failure", zap.Error(err))
		return nil, nil, err
	}
	
	d := &Data{
		db:     db.Debug(),
		redis:  redis,
		logger: logger,
	}
	
	cleanup := func() {
		err = sqlDB.Close()
		if err != nil {
			logger.Error("closing the MySQL resources failure", zap.Error(err))
		} else {
			logger.Info("closing the MySQL resources is ok")
		}
		
		err = redis.Close()
		if err != nil {
			logger.Error("closing the redis resources failure", zap.Error(err))
		} else {
			logger.Error("closing the redis resources is ok", zap.Error(err))
		}
		
		logger.Info("closing the data resources is over")
	}
	
	return d, cleanup, nil
}
