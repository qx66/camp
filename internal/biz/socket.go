package biz

import (
	"fmt"
	"go.uber.org/zap"
	"net"
	"time"
)

type SocketClientUseCase struct {
	logger *zap.Logger
}

func NewSocketClientUseCase(logger *zap.Logger) *SocketClientUseCase {
	return &SocketClientUseCase{
		logger: logger,
	}
}

func (socketClientUseCase *SocketClientUseCase) Socket(addr string, port int, timeout int) (net.Conn, error) {
	return net.DialTimeout("tcp", fmt.Sprintf("%s:%d", addr, port), time.Duration(timeout)*time.Second)
}
