package service

import (
	"github.com/google/wire"
	"github.com/qx66/camp/internal/biz"
	"go.uber.org/zap"
)

var ProviderSet = wire.NewSet(NewUseCase)

type UseCase struct {
	messageUseCase  *biz.MessageUseCase
	instructUseCase *biz.InstructUseCase
	instanceUseCase *biz.InstanceUseCase
	logger          *zap.Logger
}

func NewUseCase(logger *zap.Logger, messageUseCase *biz.MessageUseCase, instructUseCase *biz.InstructUseCase, instanceUseCase *biz.InstanceUseCase) *UseCase {
	return &UseCase{
		messageUseCase:  messageUseCase,
		instructUseCase: instructUseCase,
		instanceUseCase: instanceUseCase,
		logger:          logger,
	}
}
