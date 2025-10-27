package gitee

import (
	"context"
	"testing"
	"time"

	"github.com/luoliwoshang/git-event-monitor/internal/models"
)

func TestGiteeClient_GemstoneMerchantRepository(t *testing.T) {
	client := NewClient()

	// 测试获取 XhyQAQ/gemstone-merchant 的事件
	events, err := client.GetEvents(context.Background(), "XhyQAQ/gemstone-merchant", "")
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}

	if len(events) == 0 {
		t.Fatal("No events returned")
	}

	t.Logf("Retrieved %d events", len(events))

	// 检查是否有 PushEvent
	var pushEvents []*models.UnifiedEvent
	for _, event := range events {
		if event.Type == "PushEvent" {
			pushEvents = append(pushEvents, event)
		}
	}

	if len(pushEvents) == 0 {
		t.Log("No PushEvent found in events, this is expected for some repositories")
		return
	}

	t.Logf("Found %d PushEvents", len(pushEvents))

	// 验证最近的 PushEvent 数据结构
	lastPush := pushEvents[0]
	if lastPush.ID == "" {
		t.Error("Event ID is empty")
	}
	if lastPush.CreatedAt == "" {
		t.Error("Event CreatedAt is empty")
	}
	if lastPush.ActorLogin == "" {
		t.Error("Actor login is empty")
	}
	if lastPush.RepoName == "" {
		t.Error("Repository name is empty")
	}

	t.Logf("Last PushEvent: ID=%s, Actor=%s, Time=%s",
		lastPush.ID, lastPush.ActorLogin, lastPush.CreatedAt)
}

func TestGiteeClient_AnalyzeWithDeadline(t *testing.T) {
	client := NewClient()

	// 使用一个过去的截止时间，验证是否能正确判断为"超过截止时间"
	// 注意：XhyQAQ/gemstone-merchant 的最后提交是 2023-07-09，所以用更早的截止时间
	pastDeadline := "2023-06-01T00:00:00Z"

	req := &models.AnalysisRequest{
		Repository: "XhyQAQ/gemstone-merchant",
		Platform:   models.PlatformGitee,
		Token:      "",
		Deadline:   pastDeadline,
	}

	result, err := client.AnalyzeCodeEvents(context.Background(), req)
	if err != nil {
		t.Fatalf("Analysis failed: %v", err)
	}

	// 如果找不到 PushEvent，这是正常的（可能仓库没有最近的推送）
	if !result.Found {
		t.Logf("No code events found, which is acceptable: %s", result.Error)
		return
	}

	if result.SubmittedBefore == nil {
		t.Fatal("Expected SubmittedBefore to be set when events are found")
	}

	// 应该返回 false（因为最后提交肯定在 2024-01-01 之后）
	if *result.SubmittedBefore {
		t.Error("Expected SubmittedBefore to be false (after deadline)")
	}

	if result.TimeDifference == "" {
		t.Error("Expected TimeDifference to be set")
	}

	t.Logf("Analysis result: Found=%v, SubmittedBefore=%v, TimeDiff=%s",
		result.Found, *result.SubmittedBefore, result.TimeDifference)
}

func TestGiteeClient_AnalyzeWithFutureDeadline(t *testing.T) {
	client := NewClient()

	// 使用一个未来的截止时间，验证是否能正确判断为"在截止时间前"
	futureDeadline := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	req := &models.AnalysisRequest{
		Repository: "XhyQAQ/gemstone-merchant",
		Platform:   models.PlatformGitee,
		Token:      "",
		Deadline:   futureDeadline,
	}

	result, err := client.AnalyzeCodeEvents(context.Background(), req)
	if err != nil {
		t.Fatalf("Analysis failed: %v", err)
	}

	// 如果找不到 PushEvent，这是正常的
	if !result.Found {
		t.Logf("No code events found, which is acceptable: %s", result.Error)
		return
	}

	if result.SubmittedBefore == nil {
		t.Fatal("Expected SubmittedBefore to be set when events are found")
	}

	// 应该返回 true（因为最后提交应该在未来时间之前）
	if !*result.SubmittedBefore {
		t.Error("Expected SubmittedBefore to be true (before deadline)")
	}

	t.Logf("Analysis result with future deadline: Found=%v, SubmittedBefore=%v, TimeDiff=%s",
		result.Found, *result.SubmittedBefore, result.TimeDifference)
}

func TestGiteeClient_Platform(t *testing.T) {
	client := NewClient()
	if client.GetPlatform() != models.PlatformGitee {
		t.Error("Expected platform to be gitee")
	}
}

func TestGiteeClient_NonExistentRepository(t *testing.T) {
	client := NewClient()

	// 测试不存在的仓库
	_, err := client.GetEvents(context.Background(), "nonexistent/repository", "")
	if err == nil {
		t.Fatal("Expected error for non-existent repository")
	}

	// 应该返回 404 错误
	expectedError := "API request failed with status 404"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}

	t.Logf("Correctly handled non-existent repository: %v", err)
}

func TestGiteeClient_AnalyzeNonExistentRepository(t *testing.T) {
	client := NewClient()

	req := &models.AnalysisRequest{
		Repository: "definitely/nonexistent",
		Platform:   models.PlatformGitee,
		Token:      "",
		Deadline:   "2024-01-01T00:00:00Z",
	}

	result, err := client.AnalyzeCodeEvents(context.Background(), req)
	if err != nil {
		t.Fatalf("AnalyzeCodeEvents should not return error, but got: %v", err)
	}

	// 应该返回未找到结果
	if result.Found {
		t.Error("Expected Found to be false for non-existent repository")
	}

	if result.EventsChecked != 0 {
		t.Errorf("Expected EventsChecked to be 0, got %d", result.EventsChecked)
	}

	if result.Error == "" {
		t.Error("Expected Error to be set for non-existent repository")
	}

	// 错误信息应该包含 404
	if !contains(result.Error, "404") {
		t.Errorf("Expected error to contain '404', got: %s", result.Error)
	}

	t.Logf("Analysis result for non-existent repo: Found=%v, Error=%s",
		result.Found, result.Error)
}

// TestGiteeClient_HasCommits_WithCommits 测试有提交记录的仓库
func TestGiteeClient_HasCommits_WithCommits(t *testing.T) {
	client := NewClient()

	// 测试 dog-can-only-be-a-dog/qci - 有 commits 的仓库
	hasCommits, err := client.HasCommits(context.Background(), "dog-can-only-be-a-dog/qci", "")
	if err != nil {
		t.Fatalf("HasCommits failed: %v", err)
	}

	if !hasCommits {
		t.Error("Expected dog-can-only-be-a-dog/qci to have commits, but got false")
	}

	t.Logf("✅ dog-can-only-be-a-dog/qci has commits: %v", hasCommits)
}

// TestGiteeClient_HasCommits_EmptyRepo 测试空仓库
func TestGiteeClient_HasCommits_EmptyRepo(t *testing.T) {
	client := NewClient()

	// 测试 hmy520/empty-repo-test - 空仓库
	hasCommits, err := client.HasCommits(context.Background(), "hmy520/empty-repo-test", "")
	if err != nil {
		t.Fatalf("HasCommits failed: %v", err)
	}

	if hasCommits {
		t.Error("Expected hmy520/empty-repo-test to be empty, but got true")
	}

	t.Logf("✅ hmy520/empty-repo-test is empty: %v", !hasCommits)
}

// TestGiteeClient_HasCommits_NonExistent 测试不存在的仓库
func TestGiteeClient_HasCommits_NonExistent(t *testing.T) {
	client := NewClient()

	// 测试不存在的仓库
	hasCommits, err := client.HasCommits(context.Background(), "definitely/does-not-exist-12345", "")
	if err != nil {
		t.Fatalf("HasCommits failed: %v", err)
	}

	if hasCommits {
		t.Error("Expected non-existent repository to return false, but got true")
	}

	t.Logf("✅ Non-existent repository returns false: %v", !hasCommits)
}

// TestGiteeClient_GetCommitCount_SmallRepo 测试小型仓库的 commit 统计
func TestGiteeClient_GetCommitCount_SmallRepo(t *testing.T) {
	client := NewClient()

	// 测试 dog-can-only-be-a-dog/qci - 有一定数量的 commits
	count, isComplete, err := client.GetCommitCount(context.Background(), "dog-can-only-be-a-dog/qci", "")
	if err != nil {
		t.Fatalf("GetCommitCount failed: %v", err)
	}

	if count == 0 {
		t.Error("Expected dog-can-only-be-a-dog/qci to have commits, but got 0")
	}

	t.Logf("✅ dog-can-only-be-a-dog/qci has %d commits (complete: %v)", count, isComplete)
}

// TestGiteeClient_GetCommitCount_EmptyRepo 测试空仓库的 commit 统计
func TestGiteeClient_GetCommitCount_EmptyRepo(t *testing.T) {
	client := NewClient()

	// 测试 hmy520/empty-repo-test - 空仓库
	count, isComplete, err := client.GetCommitCount(context.Background(), "hmy520/empty-repo-test", "")
	if err != nil {
		t.Fatalf("GetCommitCount failed: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected empty repo to have 0 commits, but got %d", count)
	}

	if !isComplete {
		t.Error("Expected isComplete to be true for empty repo")
	}

	t.Logf("✅ Empty repo has %d commits (complete: %v)", count, isComplete)
}

// TestGiteeClient_GetCommitCount_NonExistent 测试不存在的仓库
func TestGiteeClient_GetCommitCount_NonExistent(t *testing.T) {
	client := NewClient()

	// 测试不存在的仓库
	count, isComplete, err := client.GetCommitCount(context.Background(), "definitely/does-not-exist-12345", "")
	if err != nil {
		t.Fatalf("GetCommitCount failed: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected non-existent repo to have 0 commits, but got %d", count)
	}

	if !isComplete {
		t.Error("Expected isComplete to be true for non-existent repo")
	}

	t.Logf("✅ Non-existent repo returns 0 commits (complete: %v)", isComplete)
}

// TestGiteeClient_GetPRCount_WithPRs 测试有PR的仓库
func TestGiteeClient_GetPRCount_WithPRs(t *testing.T) {
	client := NewClient()

	// 测试 OpenCloudOS/OpenCloudOS-Kernel - 应该有很多 PRs
	count, isComplete, err := client.GetPRCount(context.Background(), "OpenCloudOS/OpenCloudOS-Kernel", "")
	if err != nil {
		t.Fatalf("GetPRCount failed: %v", err)
	}

	if count == 0 {
		t.Error("Expected OpenCloudOS/OpenCloudOS-Kernel to have PRs, but got 0")
	}

	t.Logf("✅ OpenCloudOS/OpenCloudOS-Kernel has %d PRs (complete: %v)", count, isComplete)
}

// TestGiteeClient_GetPRCount_NoPRs 测试没有PR的仓库
func TestGiteeClient_GetPRCount_NoPRs(t *testing.T) {
	client := NewClient()

	// 测试 hmy520/empty-repo-test - 空仓库，没有PRs
	count, isComplete, err := client.GetPRCount(context.Background(), "hmy520/empty-repo-test", "")
	if err != nil {
		t.Fatalf("GetPRCount failed: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected empty repo to have 0 PRs, but got %d", count)
	}

	t.Logf("✅ Empty repo has %d PRs (complete: %v)", count, isComplete)
}

// TestGiteeClient_GetPRCount_NonExistent 测试不存在的仓库
func TestGiteeClient_GetPRCount_NonExistent(t *testing.T) {
	client := NewClient()

	// 测试不存在的仓库
	count, isComplete, err := client.GetPRCount(context.Background(), "definitely/does-not-exist-12345", "")
	if err != nil {
		t.Fatalf("GetPRCount failed: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected non-existent repo to have 0 PRs, but got %d", count)
	}

	t.Logf("✅ Non-existent repo returns 0 PRs (complete: %v)", isComplete)
}

// TestGiteeClient_GetPRStats 测试 PR 状态统计
func TestGiteeClient_GetPRStats(t *testing.T) {
	client := NewClient()

	// 测试 OpenCloudOS/OpenCloudOS-Kernel - 有三种状态的 PRs
	stats, isComplete, err := client.GetPRStats(context.Background(), "OpenCloudOS/OpenCloudOS-Kernel", "")
	if err != nil {
		t.Fatalf("GetPRStats failed: %v", err)
	}

	if stats.Total == 0 {
		t.Error("Expected OpenCloudOS/OpenCloudOS-Kernel to have PRs, but got 0")
	}

	t.Logf("✅ OpenCloudOS/OpenCloudOS-Kernel PR stats: Total=%d, Open=%d, Closed=%d, Merged=%d (complete: %v)",
		stats.Total, stats.Open, stats.Closed, stats.Merged, isComplete)

	// 验证统计的一致性
	if stats.Total != stats.Open+stats.Closed+stats.Merged {
		t.Errorf("PR count mismatch: Total=%d, but Open+Closed+Merged=%d",
			stats.Total, stats.Open+stats.Closed+stats.Merged)
	}

	// 验证三种状态都存在（必须都有数据）
	if stats.Open == 0 {
		t.Error("Expected to have open PRs, but got 0")
	}
	if stats.Closed == 0 {
		t.Error("Expected to have closed PRs, but got 0")
	}
	if stats.Merged == 0 {
		t.Error("Expected to have merged PRs, but got 0")
	}
}

// TestGiteeClient_GetPRList_AuthorInfo 测试 PR 作者信息解析
func TestGiteeClient_GetPRList_AuthorInfo(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// 使用 OpenCloudOS/OpenCloudOS-Kernel - 一个有很多 PR 的公开仓库
	repo := "OpenCloudOS/OpenCloudOS-Kernel"

	prs, _, err := client.getPRList(ctx, repo, "")
	if err != nil {
		t.Fatalf("Failed to get PR list: %v", err)
	}

	if len(prs) == 0 {
		t.Fatal("No PRs found in test repository - this should not happen for OpenCloudOS/OpenCloudOS-Kernel")
	}

	pr := prs[0]
	t.Logf("✅ Gitee PR Author Info:")
	t.Logf("  Login: %s", pr.Author.Login)
	t.Logf("  Name: %s", pr.Author.Name)
	t.Logf("  Email: %s", pr.Author.Email)
	t.Logf("  PR State: %s", pr.State)

	// 验证第一个 PR 的固定数据（基于当前 API 返回的真实数据）
	// OpenCloudOS/OpenCloudOS-Kernel 的第一个 PR（按创建时间倒序）的作者是 guzitao
	expectedLogin := "guzitao"
	if pr.Author.Login != expectedLogin {
		t.Errorf("Expected first PR author login to be %s, but got %s", expectedLogin, pr.Author.Login)
	}

	// Gitee 通常会返回 Name 字段
	expectedName := "guzitao"
	if pr.Author.Name != expectedName {
		t.Errorf("Expected first PR author name to be %s, but got %s", expectedName, pr.Author.Name)
	}

	// Email 通常为空，只记录
	if pr.Author.Email != "" {
		t.Logf("Note: Author Email is not empty: %s (Gitee API usually doesn't return this)", pr.Author.Email)
	}

	// 验证 PR 状态也被正确解析
	if pr.State == "" {
		t.Error("Expected PR to have a state, but got empty string")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}