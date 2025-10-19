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

	// HasCommits 检查仓库是否有提交记录
	// 通过调用 /commits API 来判断仓库是否为空
	// 返回 true 表示仓库有代码提交，false 表示空仓库
	HasCommits(ctx context.Context, repo string, token string) (bool, error)

	// GetCommitCount 获取仓库的 commit 总数
	// 通过分页遍历 commits API 统计提交数量，最多统计 1000 个
	// 返回值：
	//   - count: commit 数量
	//   - isComplete: 是否完整统计（true=完整统计，false=达到1000上限）
	//   - error: 错误信息
	GetCommitCount(ctx context.Context, repo string, token string) (count int, isComplete bool, err error)

	// GetPRCount 获取仓库的 PR/MR 总数
	// 通过分页遍历 pull requests API 统计 PR 数量，最多统计 1000 个
	// 返回值：
	//   - count: PR 数量
	//   - isComplete: 是否完整统计（true=完整统计，false=达到1000上限）
	//   - error: 错误信息
	GetPRCount(ctx context.Context, repo string, token string) (count int, isComplete bool, err error)

	// GetPRStats 获取仓库的 PR 统计信息（按状态分类）
	// 通过分页遍历 pull requests API 统计各状态 PR 数量，最多统计 1000 个
	// 返回值：
	//   - stats: PR 统计信息（总数、open、closed、merged）
	//   - isComplete: 是否完整统计（true=完整统计，false=达到1000上限）
	//   - error: 错误信息
	GetPRStats(ctx context.Context, repo string, token string) (stats *models.PRStats, isComplete bool, err error)

	// GetPlatform 获取平台类型
	GetPlatform() models.Platform
}

// RequestOptions API 请求选项
type RequestOptions struct {
	Token   string
	PerPage int
}