package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/luoliwoshang/git-event-monitor/internal/models"
)

const (
	maxPages = 10  // 最多获取 10 页
	perPage  = 100 // 每页 100 条记录
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

// getPaginatedList 通用的分页获取函数
// 用于获取 GitHub API 的分页数据，最多获取 1000 条记录（10页 × 100条）
// 返回值：
//   - results: 原始数据列表
//   - isComplete: 是否完整获取（true=完整，false=达到上限）
//   - error: 错误信息
func (c *Client) getPaginatedList(ctx context.Context, path string, token string, params url.Values) ([]map[string]interface{}, bool, error) {
	var results []map[string]interface{}
	page := 1

	for page <= maxPages {
		// 设置分页参数
		queryParams := url.Values{}
		for k, v := range params {
			queryParams[k] = v
		}
		queryParams.Set("per_page", fmt.Sprintf("%d", perPage))
		queryParams.Set("page", fmt.Sprintf("%d", page))

		// 构建完整 URL
		fullURL := fmt.Sprintf("%s%s?%s", c.baseURL, path, queryParams.Encode())

		// 创建 HTTP 请求
		req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
		if err != nil {
			return nil, false, fmt.Errorf("创建请求失败: %w", err)
		}

		// 设置请求头
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		// 发送 HTTP 请求
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, false, fmt.Errorf("请求失败: %w", err)
		}

		// 检查状态码
		if resp.StatusCode == 409 || resp.StatusCode == 404 {
			// 空仓库或不存在
			resp.Body.Close()
			return []map[string]interface{}{}, true, nil
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			return nil, false, fmt.Errorf("GitHub API返回异常状态码: %d", resp.StatusCode)
		}

		// 解析响应
		var pageData []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&pageData); err != nil {
			resp.Body.Close()
			return nil, false, fmt.Errorf("解析响应失败: %w", err)
		}
		resp.Body.Close()

		// 如果没有数据了，说明已经到最后一页
		if len(pageData) == 0 {
			return results, true, nil
		}

		results = append(results, pageData...)

		// 如果返回的数据少于 perPage，说明这是最后一页
		if len(pageData) < perPage {
			return results, true, nil
		}

		page++
	}

	// 达到最大页数限制，返回非完整列表
	return results, false, nil
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

// getCommitList 获取GitHub仓库的commit列表（内部方法）
// 通过分页遍历commits API获取提交列表，最多获取1000个
// 返回值：
//   - commits: commit列表
//   - isComplete: 是否完整获取（true=完整，false=达到1000上限）
//   - error: 错误信息
func (c *Client) getCommitList(ctx context.Context, repo string, token string) ([]models.Commit, bool, error) {
	path := fmt.Sprintf("/repos/%s/commits", repo)
	params := url.Values{}

	rawCommits, isComplete, err := c.getPaginatedList(ctx, path, token, params)
	if err != nil {
		return nil, false, err
	}

	// 转换为 Commit 结构（目前为空结构体）
	var commits []models.Commit
	for range rawCommits {
		commits = append(commits, models.Commit{})
	}

	return commits, isComplete, nil
}

// GetCommitCount 获取GitHub仓库的commit总数
// 通过分页遍历commits API统计提交数量，最多统计1000个
// 返回值：
//   - count: commit数量
//   - isComplete: 是否完整统计（true=完整统计，false=达到1000上限）
//   - error: 错误信息
func (c *Client) GetCommitCount(ctx context.Context, repo string, token string) (int, bool, error) {
	commits, isComplete, err := c.getCommitList(ctx, repo, token)
	if err != nil {
		return 0, false, err
	}
	return len(commits), isComplete, nil
}

// getPRList 获取GitHub仓库的PR列表（内部方法）
// 通过分页遍历pull requests API获取PR列表，最多获取1000个
// 返回值：
//   - prs: PR列表
//   - isComplete: 是否完整获取（true=完整，false=达到1000上限）
//   - error: 错误信息
func (c *Client) getPRList(ctx context.Context, repo string, token string) ([]models.PullRequest, bool, error) {
	path := fmt.Sprintf("/repos/%s/pulls", repo)
	params := url.Values{}
	params.Set("state", "all")

	rawPRs, isComplete, err := c.getPaginatedList(ctx, path, token, params)
	if err != nil {
		return nil, false, err
	}

	// 转换为 PullRequest 结构并解析状态
	var prs []models.PullRequest
	for _, rawPR := range rawPRs {
		pr := models.PullRequest{}

		// 解析状态
		state, _ := rawPR["state"].(string)
		mergedAt, _ := rawPR["merged_at"]

		// GitHub API 的 state 只有 "open" 和 "closed"
		// 需要检查 merged_at 来区分 closed 和 merged
		if state == "open" {
			pr.State = models.PRStateOpen
		} else if mergedAt != nil {
			pr.State = models.PRStateMerged
		} else {
			pr.State = models.PRStateClosed
		}

		// 解析作者信息
		if userMap, ok := rawPR["user"].(map[string]interface{}); ok {
			if login, ok := userMap["login"].(string); ok {
				pr.Author.Login = login
			}
			if name, ok := userMap["name"].(string); ok {
				pr.Author.Name = name
			}
			if email, ok := userMap["email"].(string); ok {
				pr.Author.Email = email
			}
		}

		prs = append(prs, pr)
	}

	return prs, isComplete, nil
}

// GetPRCount 获取GitHub仓库的PR总数
// 通过分页遍历pull requests API统计PR数量，最多统计1000个
// 返回值：
//   - count: PR数量
//   - isComplete: 是否完整统计（true=完整统计，false=达到1000上限）
//   - error: 错误信息
func (c *Client) GetPRCount(ctx context.Context, repo string, token string) (int, bool, error) {
	prs, isComplete, err := c.getPRList(ctx, repo, token)
	if err != nil {
		return 0, false, err
	}
	return len(prs), isComplete, nil
}

// GetPRStats 获取GitHub仓库的PR统计信息（按状态分类）
// 通过分页遍历pull requests API统计各状态PR数量，最多统计1000个
// 返回值：
//   - stats: PR统计信息（总数、open、closed、merged）
//   - isComplete: 是否完整统计（true=完整统计，false=达到1000上限）
//   - error: 错误信息
func (c *Client) GetPRStats(ctx context.Context, repo string, token string) (*models.PRStats, bool, error) {
	prs, isComplete, err := c.getPRList(ctx, repo, token)
	if err != nil {
		return nil, false, err
	}

	stats := &models.PRStats{
		Total: len(prs),
	}

	// 统计各状态的PR数量
	for _, pr := range prs {
		switch pr.State {
		case models.PRStateOpen:
			stats.Open++
		case models.PRStateClosed:
			stats.Closed++
		case models.PRStateMerged:
			stats.Merged++
		}
	}

	return stats, isComplete, nil
}