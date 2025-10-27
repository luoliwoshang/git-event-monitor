package github

import (
	"context"
	"testing"
	"time"

	"github.com/luoliwoshang/git-event-monitor/internal/models"
)

func TestGitHubClient_VSCodeRepository(t *testing.T) {
	client := NewClient()

	// 测试获取 microsoft/vscode 的事件
	events, err := client.GetEvents(context.Background(), "microsoft/vscode", "")
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
		t.Fatal("No PushEvent found in events")
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

func TestGitHubClient_AnalyzeWithDeadline(t *testing.T) {
	client := NewClient()

	// 使用一个过去的截止时间，验证是否能正确判断为"超过截止时间"
	pastDeadline := "2024-01-01T00:00:00Z"

	req := &models.AnalysisRequest{
		Repository: "microsoft/vscode",
		Platform:   models.PlatformGitHub,
		Token:      "",
		Deadline:   pastDeadline,
	}

	result, err := client.AnalyzeCodeEvents(context.Background(), req)
	if err != nil {
		t.Fatalf("Analysis failed: %v", err)
	}

	if !result.Found {
		t.Fatal("Expected to find code events")
	}

	if result.SubmittedBefore == nil {
		t.Fatal("Expected SubmittedBefore to be set")
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

func TestGitHubClient_AnalyzeWithFutureDeadline(t *testing.T) {
	client := NewClient()

	// 使用一个未来的截止时间，验证是否能正确判断为"在截止时间前"
	futureDeadline := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	req := &models.AnalysisRequest{
		Repository: "microsoft/vscode",
		Platform:   models.PlatformGitHub,
		Token:      "",
		Deadline:   futureDeadline,
	}

	result, err := client.AnalyzeCodeEvents(context.Background(), req)
	if err != nil {
		t.Fatalf("Analysis failed: %v", err)
	}

	if !result.Found {
		t.Fatal("Expected to find code events")
	}

	if result.SubmittedBefore == nil {
		t.Fatal("Expected SubmittedBefore to be set")
	}

	// 应该返回 true（因为最后提交应该在未来时间之前）
	if !*result.SubmittedBefore {
		t.Error("Expected SubmittedBefore to be true (before deadline)")
	}

	t.Logf("Analysis result with future deadline: Found=%v, SubmittedBefore=%v, TimeDiff=%s",
		result.Found, *result.SubmittedBefore, result.TimeDifference)
}

func TestGitHubClient_NonExistentRepository(t *testing.T) {
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

func TestGitHubClient_AnalyzeNonExistentRepository(t *testing.T) {
	client := NewClient()

	req := &models.AnalysisRequest{
		Repository: "definitely/nonexistent",
		Platform:   models.PlatformGitHub,
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

// TestGitHubClient_HasCommits_WithCommits 测试有提交记录的仓库
func TestGitHubClient_HasCommits_WithCommits(t *testing.T) {
	client := NewClient()

	// 测试 goplus/xgo - 活跃项目，肯定有 commits
	hasCommits, err := client.HasCommits(context.Background(), "goplus/xgo", "")
	if err != nil {
		t.Fatalf("HasCommits failed: %v", err)
	}

	if !hasCommits {
		t.Error("Expected goplus/xgo to have commits, but got false")
	}

	t.Logf("✅ goplus/xgo has commits: %v", hasCommits)
}

// TestGitHubClient_HasCommits_EmptyRepo 测试空仓库
func TestGitHubClient_HasCommits_EmptyRepo(t *testing.T) {
	client := NewClient()

	// 测试 luoliwoshang/empty-repo-test - 空仓库
	hasCommits, err := client.HasCommits(context.Background(), "luoliwoshang/empty-repo-test", "")
	if err != nil {
		t.Fatalf("HasCommits failed: %v", err)
	}

	if hasCommits {
		t.Error("Expected luoliwoshang/empty-repo-test to be empty, but got true")
	}

	t.Logf("✅ luoliwoshang/empty-repo-test is empty: %v", !hasCommits)
}

// TestGitHubClient_HasCommits_NonExistent 测试不存在的仓库
func TestGitHubClient_HasCommits_NonExistent(t *testing.T) {
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

// TestGitHubClient_GetCommitCount_SmallRepo 测试小型仓库的 commit 统计
func TestGitHubClient_GetCommitCount_SmallRepo(t *testing.T) {
	client := NewClient()

	// 测试 goplus/xgo - 有一定数量的 commits
	count, isComplete, err := client.GetCommitCount(context.Background(), "goplus/xgo", "")
	if err != nil {
		t.Fatalf("GetCommitCount failed: %v", err)
	}

	if count == 0 {
		t.Error("Expected goplus/xgo to have commits, but got 0")
	}

	t.Logf("✅ goplus/xgo has %d commits (complete: %v)", count, isComplete)
}

// TestGitHubClient_GetCommitCount_EmptyRepo 测试空仓库的 commit 统计
func TestGitHubClient_GetCommitCount_EmptyRepo(t *testing.T) {
	client := NewClient()

	// 测试 luoliwoshang/empty-repo-test - 空仓库
	count, isComplete, err := client.GetCommitCount(context.Background(), "luoliwoshang/empty-repo-test", "")
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

// TestGitHubClient_GetCommitCount_NonExistent 测试不存在的仓库
func TestGitHubClient_GetCommitCount_NonExistent(t *testing.T) {
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

// TestGitHubClient_GetPRCount_WithPRs 测试有PR的仓库
func TestGitHubClient_GetPRCount_WithPRs(t *testing.T) {
	client := NewClient()

	// 测试 goplus/xgo - 应该有一些 PRs
	count, isComplete, err := client.GetPRCount(context.Background(), "goplus/xgo", "")
	if err != nil {
		t.Fatalf("GetPRCount failed: %v", err)
	}

	if count == 0 {
		t.Error("Expected goplus/xgo to have PRs, but got 0")
	}

	t.Logf("✅ goplus/xgo has %d PRs (complete: %v)", count, isComplete)
}

// TestGitHubClient_GetPRCount_NoPRs 测试没有PR的仓库
func TestGitHubClient_GetPRCount_NoPRs(t *testing.T) {
	client := NewClient()

	// 测试 luoliwoshang/empty-repo-test - 空仓库，没有PRs
	count, isComplete, err := client.GetPRCount(context.Background(), "luoliwoshang/empty-repo-test", "")
	if err != nil {
		t.Fatalf("GetPRCount failed: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected empty repo to have 0 PRs, but got %d", count)
	}

	t.Logf("✅ Empty repo has %d PRs (complete: %v)", count, isComplete)
}

// TestGitHubClient_GetPRCount_NonExistent 测试不存在的仓库
func TestGitHubClient_GetPRCount_NonExistent(t *testing.T) {
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

// TestGitHubClient_GetPRStats 测试 PR 状态统计
func TestGitHubClient_GetPRStats(t *testing.T) {
	client := NewClient()

	// 测试 goplus/llpkg - 有三种状态的 PRs
	stats, isComplete, err := client.GetPRStats(context.Background(), "goplus/llpkg", "")
	if err != nil {
		t.Fatalf("GetPRStats failed: %v", err)
	}

	if stats.Total == 0 {
		t.Error("Expected goplus/llpkg to have PRs, but got 0")
	}

	t.Logf("✅ goplus/llpkg PR stats: Total=%d, Open=%d, Closed=%d, Merged=%d (complete: %v)",
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

// TestGitHubClient_Platform 测试平台类型
func TestGitHubClient_Platform(t *testing.T) {
	client := NewClient()
	if client.GetPlatform() != models.PlatformGitHub {
		t.Error("Expected platform to be GitHub")
	}
}

// TestGitHubClient_GetPRList_AuthorInfo 测试 PR 作者信息解析
func TestGitHubClient_GetPRList_AuthorInfo(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// 使用 luoliwoshang/prompt-front-overflow - 有 79 个 PR 的仓库
	// API 返回顺序：从新到旧（newest first）
	repo := "luoliwoshang/prompt-front-overflow"

	prs, _, err := client.getPRList(ctx, repo, "")
	if err != nil {
		t.Fatalf("Failed to get PR list: %v", err)
	}

	if len(prs) < 79 {
		t.Fatalf("Expected at least 79 PRs, but got %d", len(prs))
	}

	t.Logf("Total PRs fetched: %d", len(prs))

	// 反转数组，使其从旧到新排列，更直观
	reversed := make([]models.PullRequest, len(prs))
	for i, pr := range prs {
		reversed[len(prs)-1-i] = pr
	}

	// 验证最早的 79 个 PR 的作者信息（从旧到新的顺序）
	// 包含 3 个真人用户：luoliwoshang (6个), MeteorsLiu (5个), 其余是 xgopilot[bot] (68个)
	oldest79 := reversed[:79]
	expectedAuthors := []string{
		"xgopilot[bot]", // PR#11
		"xgopilot[bot]", // PR#12
		"xgopilot[bot]", // PR#17
		"xgopilot[bot]", // PR#18
		"xgopilot[bot]", // PR#19
		"xgopilot[bot]", // PR#20
		"xgopilot[bot]", // PR#23
		"xgopilot[bot]", // PR#25
		"xgopilot[bot]", // PR#27
		"xgopilot[bot]", // PR#28
		"xgopilot[bot]", // PR#29
		"xgopilot[bot]", // PR#31
		"xgopilot[bot]", // PR#33
		"xgopilot[bot]", // PR#38
		"xgopilot[bot]", // PR#47
		"xgopilot[bot]", // PR#49
		"xgopilot[bot]", // PR#51
		"xgopilot[bot]", // PR#52
		"xgopilot[bot]", // PR#53
		"xgopilot[bot]", // PR#55
		"xgopilot[bot]", // PR#57
		"luoliwoshang",  // PR#58
		"xgopilot[bot]", // PR#61
		"xgopilot[bot]", // PR#62
		"xgopilot[bot]", // PR#66
		"xgopilot[bot]", // PR#67
		"xgopilot[bot]", // PR#68
		"xgopilot[bot]", // PR#70
		"xgopilot[bot]", // PR#72
		"xgopilot[bot]", // PR#74
		"xgopilot[bot]", // PR#76
		"xgopilot[bot]", // PR#78
		"xgopilot[bot]", // PR#79
		"xgopilot[bot]", // PR#80
		"MeteorsLiu",    // PR#81
		"xgopilot[bot]", // PR#83
		"MeteorsLiu",    // PR#84
		"xgopilot[bot]", // PR#85
		"luoliwoshang",  // PR#86
		"xgopilot[bot]", // PR#88
		"luoliwoshang",  // PR#89
		"xgopilot[bot]", // PR#91
		"xgopilot[bot]", // PR#92
		"xgopilot[bot]", // PR#94
		"xgopilot[bot]", // PR#96
		"xgopilot[bot]", // PR#98
		"xgopilot[bot]", // PR#100
		"xgopilot[bot]", // PR#101
		"xgopilot[bot]", // PR#103
		"luoliwoshang",  // PR#105
		"luoliwoshang",  // PR#106
		"luoliwoshang",  // PR#108
		"xgopilot[bot]", // PR#112
		"xgopilot[bot]", // PR#113
		"xgopilot[bot]", // PR#120
		"xgopilot[bot]", // PR#122
		"xgopilot[bot]", // PR#124
		"luoliwoshang",  // PR#125
		"MeteorsLiu",    // PR#126
		"xgopilot[bot]", // PR#129
		"xgopilot[bot]", // PR#131
		"MeteorsLiu",    // PR#132
		"xgopilot[bot]", // PR#133
		"xgopilot[bot]", // PR#134
		"xgopilot[bot]", // PR#135
		"xgopilot[bot]", // PR#137
		"MeteorsLiu",    // PR#138
		"xgopilot[bot]", // PR#139
		"xgopilot[bot]", // PR#141
		"xgopilot[bot]", // PR#143
		"xgopilot[bot]", // PR#144
		"xgopilot[bot]", // PR#146
		"xgopilot[bot]", // PR#148
		"xgopilot[bot]", // PR#150
		"xgopilot[bot]", // PR#153
		"xgopilot[bot]", // PR#155
		"xgopilot[bot]", // PR#156
		"xgopilot[bot]", // PR#157
		"xgopilot[bot]", // PR#158
	}

	for i, pr := range oldest79 {
		expected := expectedAuthors[i]

		// 验证作者 Login 匹配预期
		if pr.Author.Login != expected {
			t.Errorf("PR index %d: Expected author to be %s, but got %s", i, expected, pr.Author.Login)
		}

		// GitHub PR 列表 API 不返回 Name 和 Email，应该为空
		if pr.Author.Name != "" {
			t.Errorf("PR index %d: Expected author name to be empty (GitHub API doesn't return it), but got %s", i, pr.Author.Name)
		}
		if pr.Author.Email != "" {
			t.Errorf("PR index %d: Expected author email to be empty (GitHub API doesn't return it), but got %s", i, pr.Author.Email)
		}

		// 验证 PR 状态
		if pr.State == "" {
			t.Errorf("PR index %d: Expected PR to have a state, but got empty string", i)
		}
	}

	t.Logf("✅ Verified first 79 PRs (from oldest): 68 by xgopilot[bot], 6 by luoliwoshang, 5 by MeteorsLiu")
}

// TestGitHubClient_GetPRList_AuthorInfo_HackathonGO 测试另一个仓库的 PR 作者信息解析
func TestGitHubClient_GetPRList_AuthorInfo_HackathonGO(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// 使用 wwcchh0123/hackathonGO - 有 48 个 PR 的仓库
	// API 返回顺序：从新到旧（newest first）
	repo := "wwcchh0123/hackathonGO"

	prs, _, err := client.getPRList(ctx, repo, "")
	if err != nil {
		t.Fatalf("Failed to get PR list: %v", err)
	}

	if len(prs) < 48 {
		t.Fatalf("Expected at least 48 PRs, but got %d", len(prs))
	}

	t.Logf("Total PRs fetched: %d", len(prs))

	// 反转数组，使其从旧到新排列
	reversed := make([]models.PullRequest, len(prs))
	for i, pr := range prs {
		reversed[len(prs)-1-i] = pr
	}

	// 验证最早的 48 个 PR 的作者信息（从旧到新的顺序）
	// 包含 4 个真人用户：CarlJi (15个), minorcell (13个), wwcchh0123 (6个), 其余是 xgopilot[bot] (14个)
	oldest48 := reversed[:48]
	expectedAuthors := []string{
		"xgopilot[bot]", // PR#4
		"minorcell",     // PR#5
		"xgopilot[bot]", // PR#8
		"xgopilot[bot]", // PR#9
		"wwcchh0123",    // PR#10
		"wwcchh0123",    // PR#11
		"CarlJi",        // PR#12
		"CarlJi",        // PR#13
		"CarlJi",        // PR#14
		"minorcell",     // PR#15
		"xgopilot[bot]", // PR#18
		"xgopilot[bot]", // PR#19
		"CarlJi",        // PR#20
		"CarlJi",        // PR#21
		"wwcchh0123",    // PR#24
		"CarlJi",        // PR#25
		"minorcell",     // PR#26
		"minorcell",     // PR#27
		"CarlJi",        // PR#28
		"wwcchh0123",    // PR#29
		"CarlJi",        // PR#31
		"minorcell",     // PR#32
		"xgopilot[bot]", // PR#35
		"xgopilot[bot]", // PR#36
		"xgopilot[bot]", // PR#39
		"xgopilot[bot]", // PR#40
		"minorcell",     // PR#41
		"CarlJi",        // PR#42
		"wwcchh0123",    // PR#43
		"xgopilot[bot]", // PR#45
		"CarlJi",        // PR#46
		"minorcell",     // PR#47
		"minorcell",     // PR#48
		"CarlJi",        // PR#50
		"xgopilot[bot]", // PR#53
		"wwcchh0123",    // PR#55
		"xgopilot[bot]", // PR#56
		"minorcell",     // PR#57
		"xgopilot[bot]", // PR#58
		"CarlJi",        // PR#59
		"xgopilot[bot]", // PR#60
		"minorcell",     // PR#61
		"minorcell",     // PR#62
		"minorcell",     // PR#63
		"CarlJi",        // PR#64
		"CarlJi",        // PR#65
		"CarlJi",        // PR#66
		"CarlJi",        // PR#68
	}

	for i, pr := range oldest48 {
		expected := expectedAuthors[i]

		// 验证作者 Login 匹配预期
		if pr.Author.Login != expected {
			t.Errorf("PR index %d: Expected author to be %s, but got %s", i, expected, pr.Author.Login)
		}

		// GitHub PR 列表 API 不返回 Name 和 Email，应该为空
		if pr.Author.Name != "" {
			t.Errorf("PR index %d: Expected author name to be empty (GitHub API doesn't return it), but got %s", i, pr.Author.Name)
		}
		if pr.Author.Email != "" {
			t.Errorf("PR index %d: Expected author email to be empty (GitHub API doesn't return it), but got %s", i, pr.Author.Email)
		}

		// 验证 PR 状态
		if pr.State == "" {
			t.Errorf("PR index %d: Expected PR to have a state, but got empty string", i)
		}
	}

	t.Logf("✅ Verified first 48 PRs (from oldest): 15 by CarlJi, 13 by minorcell, 6 by wwcchh0123, 14 by xgopilot[bot]")
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