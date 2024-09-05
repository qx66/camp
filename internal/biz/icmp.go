package biz

import (
	"github.com/prometheus-community/pro-bing"
	"go.uber.org/zap"
	"time"
)

type IcmpClientUseCase struct {
	logger *zap.Logger
}

func NewIcmpClientUseCase(logger *zap.Logger) *IcmpClientUseCase {
	return &IcmpClientUseCase{
		logger: logger,
	}
}

func (icmpClientUseCase *IcmpClientUseCase) Icmp(addr string, count int) (*probing.Statistics, error) {
	pinger, err := probing.NewPinger(addr)
	
	if err != nil {
		
		return nil, err
	}
	
	pinger.Count = count
	pinger.Timeout = 5 * time.Second
	
	err = pinger.Run()
	if err != nil {
		return nil, err
	}
	
	return pinger.Statistics(), nil
}
