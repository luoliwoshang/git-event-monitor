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