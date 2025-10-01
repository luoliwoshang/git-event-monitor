package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/luoliwoshang/git-event-monitor/internal/models"
)

// Client GitHub API 客户端
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient 创建新的 GitHub 客户端
func NewClient() *Client {
	return &Client{
		baseURL: "https://api.github.com",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetPlatform 获取平台类型
func (c *Client) GetPlatform() models.Platform {
	return models.PlatformGitHub
}

// GetEvents 获取仓库事件列表
func (c *Client) GetEvents(ctx context.Context, repo string, token string) ([]*models.UnifiedEvent, error) {
	url := fmt.Sprintf("%s/repos/%s/events", c.baseURL, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "git-event-monitor/1.0")

	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	// 设置查询参数
	q := req.URL.Query()
	q.Set("per_page", "100")
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var githubEvents []models.GitHubEvent
	if err := json.NewDecoder(resp.Body).Decode(&githubEvents); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// 转换为统一事件格式
	var events []*models.UnifiedEvent
	for _, event := range githubEvents {
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
		result.Error = fmt.Sprintf("No code submission events found in the last %d repository events", len(events))
		return result, nil
	}

	// 获取最近的代码事件
	lastEvent := codeEvents[0]
	result.LastCodeEvent = lastEvent
	result.EventDescription = fmt.Sprintf("Latest %s (%s)", lastEvent.Type, lastEvent.CreatedAt)

	// 如果提供了截止时间，检查合规性
	if req.Deadline != "" {
		deadline, err := time.Parse(time.RFC3339, req.Deadline)
		if err != nil {
			result.Error = fmt.Sprintf("Invalid deadline format: %s", err.Error())
			return result, nil
		}

		eventTime, err := time.Parse(time.RFC3339, lastEvent.CreatedAt)
		if err != nil {
			result.Error = fmt.Sprintf("Invalid event time format: %s", err.Error())
			return result, nil
		}

		isBeforeDeadline := eventTime.Before(deadline) || eventTime.Equal(deadline)
		result.SubmittedBefore = &isBeforeDeadline

		// 计算时间差
		timeDiff := deadline.Sub(eventTime)
		if timeDiff > 0 {
			result.TimeDifference = fmt.Sprintf("%s before deadline", formatDuration(timeDiff))
		} else {
			result.TimeDifference = fmt.Sprintf("%s after deadline", formatDuration(-timeDiff))
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
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		if minutes > 0 {
			return fmt.Sprintf("%d hours %d minutes", hours, minutes)
		}
		return fmt.Sprintf("%d hours", hours)
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	if hours > 0 {
		return fmt.Sprintf("%d days %d hours", days, hours)
	}
	return fmt.Sprintf("%d days", days)
}

// HasCommits 检查GitHub仓库是否有提交记录
// 通过调用GitHub Commits API来判断仓库是否为空
// 返回值：
//   - true: 仓库有代码提交
//   - false: 仓库为空（无提交记录）
//   - error: API调用失败或其他错误
func (c *Client) HasCommits(ctx context.Context, repo string, token string) (bool, error) {
	// 构建API URL，只请求第一个commit来减少开销
	url := fmt.Sprintf("%s/repos/%s/commits?per_page=1", c.baseURL, repo)

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	// 如果提供了token，添加认证头
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
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
	case 409:
		// 状态码409表示仓库为空（Git Repository is empty）
		return false, nil
	case 404:
		// 状态码404表示仓库不存在或无权限访问
		// 在这种情况下，我们认为是没有提交记录
		return false, nil
	default:
		// 其他状态码表示API调用出现异常
		return false, fmt.Errorf("GitHub API返回异常状态码: %d", resp.StatusCode)
	}
}