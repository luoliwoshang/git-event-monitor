package main

import (
	"context"
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/luoliwoshang/git-event-monitor/internal/api"
	"github.com/luoliwoshang/git-event-monitor/internal/api/gitee"
	"github.com/luoliwoshang/git-event-monitor/internal/api/github"
)

// TestGiteeRepository 测试 Gitee 仓库统计功能
// 仓库: gitee.com/hmy520/test-qiniu
// 预期结果: 3 个 commit, 1 个 open PR, 1 个 merged PR, 1 个 closed PR
func TestGiteeRepository(t *testing.T) {
	// 读取 Gitee Token
	giteeToken := os.Getenv("GITEE_TOKEN")
	if giteeToken == "" {
		t.Fatal("GITEE_TOKEN not set - test requires valid token")
	}

	// 创建 Gitee 客户端
	client := gitee.NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repo := "hmy520/test-qiniu"

	t.Run("GetCommitCount", func(t *testing.T) {
		count, isComplete, err := client.GetCommitCount(ctx, repo, giteeToken)
		if err != nil {
			t.Fatalf("GetCommitCount failed: %v", err)
		}

		t.Logf("Commit count: %d, isComplete: %v", count, isComplete)

		// 验证 commit 数量
		expectedCount := 3
		if count != expectedCount {
			t.Errorf("Expected %d commits, got %d", expectedCount, count)
		}

		// 由于只有 3 个 commit，应该是完整统计
		if !isComplete {
			t.Errorf("Expected isComplete=true for small repository, got false")
		}
	})

	t.Run("GetPRStats", func(t *testing.T) {
		stats, isComplete, err := client.GetPRStats(ctx, repo, giteeToken)
		if err != nil {
			t.Fatalf("GetPRStats failed: %v", err)
		}

		t.Logf("PR Stats: Total=%d, Open=%d, Closed=%d, Merged=%d, isComplete=%v",
			stats.Total, stats.Open, stats.Closed, stats.Merged, isComplete)

		// 验证 PR 总数
		expectedTotal := 3
		if stats.Total != expectedTotal {
			t.Errorf("Expected total=%d PRs, got %d", expectedTotal, stats.Total)
		}

		// 验证各状态的 PR 数量
		expectedOpen := 1
		if stats.Open != expectedOpen {
			t.Errorf("Expected %d open PRs, got %d", expectedOpen, stats.Open)
		}

		expectedMerged := 1
		if stats.Merged != expectedMerged {
			t.Errorf("Expected %d merged PRs, got %d", expectedMerged, stats.Merged)
		}

		expectedClosed := 1
		if stats.Closed != expectedClosed {
			t.Errorf("Expected %d closed PRs, got %d", expectedClosed, stats.Closed)
		}

		// 验证状态总和等于总数
		if stats.Open+stats.Closed+stats.Merged != stats.Total {
			t.Errorf("Status counts don't add up: %d+%d+%d != %d",
				stats.Open, stats.Closed, stats.Merged, stats.Total)
		}

		// 由于只有 3 个 PR，应该是完整统计
		if !isComplete {
			t.Errorf("Expected isComplete=true for small repository, got false")
		}
	})
}

// TestProcessCSV 测试端到端的 CSV 处理流程
func TestProcessCSV(t *testing.T) {
	// 读取环境变量中的 token
	giteeToken := os.Getenv("GITEE_TOKEN")
	githubToken := os.Getenv("GITHUB_TOKEN")

	if giteeToken == "" && githubToken == "" {
		t.Fatal("No tokens set (GITEE_TOKEN or GITHUB_TOKEN) - test requires valid tokens")
	}

	t.Run("Gitee_Repository", func(t *testing.T) {
		if giteeToken == "" {
			t.Fatal("GITEE_TOKEN not set - test requires valid token")
		}

		// 读取测试输入文件
		inputFile := filepath.Join("testdata", "gitee_test_input.csv")
		records, err := readCSVFile(inputFile)
		if err != nil {
			t.Fatalf("Failed to read input CSV: %v", err)
		}

		if len(records) < 2 {
			t.Fatalf("Input CSV must have at least header and one data row")
		}

		// 处理 CSV
		processedRecords := processCSVRecords(t, records, giteeToken, githubToken)

		// 读取期望的输出文件
		expectedFile := filepath.Join("testdata", "gitee_test_expected.csv")
		expectedRecords, err := readCSVFile(expectedFile)
		if err != nil {
			t.Fatalf("Failed to read expected CSV: %v", err)
		}

		// 对比结果
		if len(processedRecords) != len(expectedRecords) {
			t.Fatalf("Row count mismatch: expected %d rows, got %d rows",
				len(expectedRecords), len(processedRecords))
		}

		for i, expectedRow := range expectedRecords {
			actualRow := processedRecords[i]
			if len(actualRow) != len(expectedRow) {
				t.Errorf("Row %d: column count mismatch: expected %d columns, got %d columns",
					i, len(expectedRow), len(actualRow))
				continue
			}

			for j, expectedCell := range expectedRow {
				actualCell := actualRow[j]
				if actualCell != expectedCell {
					t.Errorf("Row %d, Column %d: expected %q, got %q",
						i, j, expectedCell, actualCell)
				}
			}
		}

		// 可选：写入实际输出文件用于调试
		outputFile := filepath.Join("testdata", "gitee_test_output.csv")
		err = writeCSVFile(outputFile, processedRecords)
		if err != nil {
			t.Fatalf("Failed to write output CSV: %v", err)
		}
		t.Logf("Output written to: %s", outputFile)

		// 清理输出文件
		defer os.Remove(outputFile)
	})
}

// processCSVRecords 处理 CSV 记录（核心业务逻辑）
func processCSVRecords(t *testing.T, records [][]string, giteeToken, githubToken string) [][]string {
	// 添加新列到表头
	headers := records[0]
	headers = append(headers, "是否可访问", "Commit数量", "PR总数", "PR-Open", "PR-Merged", "PR-Closed")

	// 为所有数据行添加空列
	for i := 1; i < len(records); i++ {
		records[i] = append(records[i], "", "", "", "", "", "")
	}

	// 更新表头
	records[0] = headers

	// 处理每一行数据
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 2 {
			continue
		}

		repoURL := record[1] // 仓库地址在第 2 列
		t.Logf("Processing row %d: %s", i, repoURL)

		// 解析仓库 URL
		platform, owner, repo := parseRepositoryURL(repoURL)
		if platform == "" {
			t.Logf("  Skipping: Cannot parse repository URL")
			continue
		}

		repoPath := owner + "/" + repo
		t.Logf("  Platform: %s, Repository: %s", platform, repoPath)

		// 创建对应平台的客户端
		var client api.Client
		var token string
		switch platform {
		case "github":
			client = github.NewClient()
			token = githubToken
		case "gitee":
			client = gitee.NewClient()
			token = giteeToken
		default:
			t.Logf("  Unsupported platform: %s", platform)
			continue
		}

		// 检查仓库是否可访问
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_, err := client.GetEvents(ctx, repoPath, token)
		cancel()

		if err != nil {
			t.Logf("  Repository not accessible: %v", err)
			record[2] = "不可访问"
			// 不可访问时，其他字段留空
			record[3] = ""
			record[4] = ""
			record[5] = ""
			record[6] = ""
			record[7] = ""
			continue
		}

		t.Logf("  Repository accessible")
		record[2] = "可访问"

		// 获取 Commit 数量
		ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
		commitCount, _, err := client.GetCommitCount(ctx2, repoPath, token)
		cancel2()

		if err != nil {
			t.Logf("  GetCommitCount error: %v", err)
			record[3] = "错误"
		} else {
			record[3] = formatInt(commitCount)
			t.Logf("  Commits: %d", commitCount)
		}

		// 获取 PR 统计
		ctx3, cancel3 := context.WithTimeout(context.Background(), 30*time.Second)
		stats, _, err := client.GetPRStats(ctx3, repoPath, token)
		cancel3()

		if err != nil {
			t.Logf("  GetPRStats error: %v", err)
			record[4] = "错误"
			record[5] = "错误"
			record[6] = "错误"
			record[7] = "错误"
		} else {
			record[4] = formatInt(stats.Total)
			record[5] = formatInt(stats.Open)
			record[6] = formatInt(stats.Merged)
			record[7] = formatInt(stats.Closed)
			t.Logf("  PRs: Total=%d, Open=%d, Merged=%d, Closed=%d",
				stats.Total, stats.Open, stats.Merged, stats.Closed)
		}
	}

	return records
}

// readCSVFile 读取 CSV 文件
func readCSVFile(filename string) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	return reader.ReadAll()
}

// writeCSVFile 写入 CSV 文件
func writeCSVFile(filename string, records [][]string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	return writer.WriteAll(records)
}

// parseRepositoryURL 解析仓库 URL（简化版，复用 csv-processor 的逻辑）
func parseRepositoryURL(url string) (platform, owner, repo string) {
	// 简单实现：检查是否包含 github 或 gitee
	if contains(url, "github") {
		// 简单解析 github.com/owner/repo
		parts := extractRepoPath(url)
		if len(parts) == 2 {
			return "github", parts[0], parts[1]
		}
	} else if contains(url, "gitee") {
		// 简单解析 gitee.com/owner/repo
		parts := extractRepoPath(url)
		if len(parts) == 2 {
			return "gitee", parts[0], parts[1]
		}
	}
	return "", "", ""
}

// contains 检查字符串是否包含子串（不区分大小写）
func contains(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return len(s) >= len(substr) && findSubstring(s, substr) != -1
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// extractRepoPath 从 URL 中提取 owner/repo
func extractRepoPath(url string) []string {
	// 移除协议前缀
	url = removePrefix(url, "https://")
	url = removePrefix(url, "http://")
	url = removePrefix(url, "git@")

	// 替换 : 为 /
	url = replace(url, ":", "/")

	// 移除 .git 后缀
	url = removeSuffix(url, ".git")

	// 分割路径
	parts := split(url, "/")

	// 期望格式: domain/owner/repo
	if len(parts) >= 3 {
		return []string{parts[1], parts[2]}
	}

	return nil
}

func removePrefix(s, prefix string) string {
	if len(s) >= len(prefix) {
		match := true
		for i := 0; i < len(prefix); i++ {
			if s[i] != prefix[i] {
				match = false
				break
			}
		}
		if match {
			return s[len(prefix):]
		}
	}
	return s
}

func removeSuffix(s, suffix string) string {
	if len(s) >= len(suffix) {
		match := true
		start := len(s) - len(suffix)
		for i := 0; i < len(suffix); i++ {
			if s[start+i] != suffix[i] {
				match = false
				break
			}
		}
		if match {
			return s[:start]
		}
	}
	return s
}

func replace(s, old, new string) string {
	result := ""
	i := 0
	for i < len(s) {
		if i <= len(s)-len(old) {
			match := true
			for j := 0; j < len(old); j++ {
				if s[i+j] != old[j] {
					match = false
					break
				}
			}
			if match {
				result += new
				i += len(old)
				continue
			}
		}
		result += string(s[i])
		i++
	}
	return result
}

func split(s, sep string) []string {
	if len(sep) == 0 {
		return []string{s}
	}

	var parts []string
	start := 0

	for i := 0; i <= len(s)-len(sep); i++ {
		match := true
		for j := 0; j < len(sep); j++ {
			if s[i+j] != sep[j] {
				match = false
				break
			}
		}
		if match {
			parts = append(parts, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}

	parts = append(parts, s[start:])
	return parts
}

// formatInt 格式化整数为字符串
func formatInt(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	digits := ""
	for n > 0 {
		digits = string(byte('0'+n%10)) + digits
		n /= 10
	}

	if negative {
		return "-" + digits
	}
	return digits
}
