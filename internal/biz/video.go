package biz

import "go.uber.org/zap"

type VideoInspectUseCase struct {
	logger *zap.Logger
}

func NewVideoInspectUseCase(logger *zap.Logger) *VideoInspectUseCase {
	return &VideoInspectUseCase{
		logger: logger,
	}
}

// 观看视频质量

func (videoInspectUseCase *VideoInspectUseCase) Video() {
	
}
