package gitee

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/luoliwoshang/git-event-monitor/internal/models"
)

// Client Gitee API 客户端
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient 创建新的 Gitee 客户端
func NewClient() *Client {
	return &Client{
		baseURL: "https://gitee.com/api/v5",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetPlatform 获取平台类型
func (c *Client) GetPlatform() models.Platform {
	return models.PlatformGitee
}

// GetEvents 获取仓库事件列表
func (c *Client) GetEvents(ctx context.Context, repo string, token string) ([]*models.UnifiedEvent, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format, expected 'owner/repo'")
	}

	url := fmt.Sprintf("%s/repos/%s/%s/events", c.baseURL, parts[0], parts[1])

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// 设置查询参数
	q := req.URL.Query()
	q.Set("limit", "100")
	if token != "" {
		q.Set("access_token", token)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var giteeEvents []models.GiteeEvent
	if err := json.NewDecoder(resp.Body).Decode(&giteeEvents); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// 转换为统一事件格式
	var events []*models.UnifiedEvent
	for _, event := range giteeEvents {
		events = append(events, event.ToUnifiedEvent())
	}

	return events, nil
}

// AnalyzeCodeEvents 分析代码提交事件
func (c *Client) AnalyzeCodeEvents(ctx context.Context, req *models.AnalysisRequest) (*models.AnalysisResult, error) {
	events, err := c.GetEvents(ctx, req.Repository, req.Token)
	if err != nil {
		return &models.AnalysisResult{
			Found:         false,
			EventsChecked: 0,
			Error:         err.Error(),
		}, nil
	}

	// 过滤代码提交事件
	var codeEvents []*models.UnifiedEvent
	for _, event := range events {
		if isCodeSubmissionEvent(event) {
			codeEvents = append(codeEvents, event)
		}
	}

	result := &models.AnalysisResult{
		Found:         len(codeEvents) > 0,
		EventsChecked: len(events),
	}

	if !result.Found {
		result.Error = fmt.Sprintf("在最近的 %d 个仓库事件中未找到代码提交事件", len(events))
		return result, nil
	}

	// 获取最近的代码事件
	lastEvent := codeEvents[0]
	result.LastCodeEvent = lastEvent
	result.EventDescription = fmt.Sprintf("最近的 %s (%s)", lastEvent.Type, lastEvent.CreatedAt)

	// 如果提供了截止时间，检查合规性
	if req.Deadline != "" {
		deadline, err := time.Parse(time.RFC3339, req.Deadline)
		if err != nil {
			result.Error = fmt.Sprintf("截止时间格式错误: %s", err.Error())
			return result, nil
		}

		eventTime, err := time.Parse(time.RFC3339, lastEvent.CreatedAt)
		if err != nil {
			result.Error = fmt.Sprintf("事件时间格式错误: %s", err.Error())
			return result, nil
		}

		isBeforeDeadline := eventTime.Before(deadline) || eventTime.Equal(deadline)
		result.SubmittedBefore = &isBeforeDeadline

		// 计算时间差
		timeDiff := deadline.Sub(eventTime)
		if timeDiff > 0 {
			result.TimeDifference = fmt.Sprintf("截止时间前 %s", formatDuration(timeDiff))
		} else {
			result.TimeDifference = fmt.Sprintf("超过截止时间 %s", formatDuration(-timeDiff))
		}
	}

	return result, nil
}

// isCodeSubmissionEvent 判断是否为代码提交事件
func isCodeSubmissionEvent(event *models.UnifiedEvent) bool {
	return event.Type == "PushEvent"
}

// formatDuration 格式化持续时间
func formatDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%d分钟", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		if minutes > 0 {
			return fmt.Sprintf("%d小时%d分钟", hours, minutes)
		}
		return fmt.Sprintf("%d小时", hours)
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	if hours > 0 {
		return fmt.Sprintf("%d天%d小时", days, hours)
	}
	return fmt.Sprintf("%d天", days)
}

// HasCommits 检查Gitee仓库是否有提交记录
// 通过调用Gitee Commits API来判断仓库是否为空
// 返回值：
//   - true: 仓库有代码提交
//   - false: 仓库为空（无提交记录）
//   - error: API调用失败或其他错误
func (c *Client) HasCommits(ctx context.Context, repo string, token string) (bool, error) {
	// 构建API URL，只请求第一个commit来减少开销
	url := fmt.Sprintf("%s/repos/%s/commits", c.baseURL, repo)

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("创建请求失败: %w", err)
	}

	// 如果提供了token，添加认证头
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	// 发送HTTP请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 根据HTTP状态码判断结果
	switch resp.StatusCode {
	case 200:
		// 状态码200表示请求成功，仓库有提交记录
		return true, nil
	case 404:
		// 状态码404表示仓库为空或不存在
		// 在这种情况下，我们认为是没有提交记录
		return false, nil
	default:
		// 其他状态码表示API调用出现异常
		return false, fmt.Errorf("Gitee API返回异常状态码: %d", resp.StatusCode)
	}
}
