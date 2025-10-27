package main

import (
	"context"
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	// GitHub Actions 不允许设置 GITHUB_ 开头的 secret，所以支持多个环境变量名
	// 优先使用 GH_TOKEN，因为 GitHub Actions 会自动设置 GITHUB_TOKEN 但权限可能不够
	giteeToken := os.Getenv("GITEE_TOKEN")
	githubToken := os.Getenv("GH_TOKEN")
	if githubToken == "" {
		githubToken = os.Getenv("GITHUB_TOKEN") // 备用变量名
	}

	if giteeToken == "" && githubToken == "" {
		t.Fatal("No tokens set (GITEE_TOKEN or GITHUB_TOKEN/GH_TOKEN) - test requires valid tokens")
	}

	t.Run("Multiple_Repositories", func(t *testing.T) {
		if giteeToken == "" && githubToken == "" {
			t.Fatal("At least one token (GITEE_TOKEN or GITHUB_TOKEN/GH_TOKEN) must be set")
		}

		// 读取测试输入文件
		inputFile := filepath.Join("testdata", "test_input.csv")
		records, err := readCSVFile(inputFile)
		if err != nil {
			t.Fatalf("Failed to read input CSV: %v", err)
		}

		if len(records) < 2 {
			t.Fatalf("Input CSV must have at least header and one data row")
		}

		// 处理 CSV
		processedRecords := processCSVRecords(t, records, giteeToken, githubToken)

		// 检查是否为生成模式
		genMode := os.Getenv("GEN_EXPECTED")
		if genMode == "true" || genMode == "1" {
			// 生成模式：将处理结果写入 expected 文件
			expectedFile := filepath.Join("testdata", "test_expected.csv")
			err = writeCSVFile(expectedFile, processedRecords)
			if err != nil {
				t.Fatalf("Failed to write expected CSV: %v", err)
			}
			t.Logf("✅ Generated expected output file: %s", expectedFile)
			t.Log("📝 Expected file has been updated. Please review and commit the changes.")
			return // 生成模式下不进行验证，直接返回
		}

		// 正常测试模式：读取期望的输出文件并对比
		expectedFile := filepath.Join("testdata", "test_expected.csv")
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
		outputFile := filepath.Join("testdata", "test_output.csv")
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
			record[3] = strconv.Itoa(commitCount)
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
			record[4] = strconv.Itoa(stats.Total)
			record[5] = strconv.Itoa(stats.Open)
			record[6] = strconv.Itoa(stats.Merged)
			record[7] = strconv.Itoa(stats.Closed)
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

// parseRepositoryURL 解析仓库 URL
func parseRepositoryURL(url string) (platform, owner, repo string) {
	// 检查是否包含 github 或 gitee
	urlLower := strings.ToLower(url)
	if strings.Contains(urlLower, "github") {
		parts := extractRepoPath(url)
		if len(parts) == 2 {
			return "github", parts[0], parts[1]
		}
	} else if strings.Contains(urlLower, "gitee") {
		parts := extractRepoPath(url)
		if len(parts) == 2 {
			return "gitee", parts[0], parts[1]
		}
	}
	return "", "", ""
}

// extractRepoPath 从 URL 中提取 owner/repo
func extractRepoPath(url string) []string {
	// 移除协议前缀
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "git@")

	// 替换 : 为 /
	url = strings.ReplaceAll(url, ":", "/")

	// 移除 .git 后缀
	url = strings.TrimSuffix(url, ".git")

	// 分割路径
	parts := strings.Split(url, "/")

	// 期望格式: domain/owner/repo
	if len(parts) >= 3 {
		return []string{parts[1], parts[2]}
	}

	return nil
}
