package biz

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestChromeDpUseCase_GetSinglePageInspect(t *testing.T) {
	logger, _ := zap.NewProduction()
	chrome := NewChromeDpClientUseCase(logger)
	
	ctx := context.Background()
	inspects, err := chrome.InspectSinglePage(ctx, "https://www.baidu.com")
	assert.Error(t, err, "ChromeDp单页面访问失败")
	
	logger.Info("inspects", zap.Any("inspects", inspects))
}
