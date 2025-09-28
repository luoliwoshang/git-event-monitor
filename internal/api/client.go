package api

import (
	"context"

	"github.com/luoliwoshang/git-event-monitor/internal/models"
)

// Client API 客户端通用接口
type Client interface {
	// GetEvents 获取仓库事件列表
	GetEvents(ctx context.Context, repo string, token string) ([]*models.UnifiedEvent, error)

	// AnalyzeCodeEvents 分析代码提交事件
	AnalyzeCodeEvents(ctx context.Context, req *models.AnalysisRequest) (*models.AnalysisResult, error)

	// GetPlatform 获取平台类型
	GetPlatform() models.Platform
}

// RequestOptions API 请求选项
type RequestOptions struct {
	Token   string
	PerPage int
}