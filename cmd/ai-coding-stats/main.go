package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/luoliwoshang/git-event-monitor/internal/api/github"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input.csv>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Environment variables:\n")
		fmt.Fprintf(os.Stderr, "  GH_TOKEN or GITHUB_TOKEN: GitHub access token (optional)\n")
		os.Exit(1)
	}

	inputFile := os.Args[1]

	// 读取输入 CSV
	records, err := readCSV(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read input CSV: %v\n", err)
		os.Exit(1)
	}

	// 处理每条记录
	processedRecords, err := processRecords(records)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to process records: %v\n", err)
		os.Exit(1)
	}

	// 输出结果
	if err := writeCSV(os.Stdout, processedRecords); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write output: %v\n", err)
		os.Exit(1)
	}
}

// readCSV 读取 CSV 文件
func readCSV(filename string) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

// writeCSV 写入 CSV 到 writer
func writeCSV(writer interface{ Write([]byte) (int, error) }, records [][]string) error {
	csvWriter := csv.NewWriter(writer.(interface {
		Write([]byte) (int, error)
	}))

	for _, record := range records {
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}

	csvWriter.Flush()
	return csvWriter.Error()
}

// processRecords 处理输入记录，计算 AI Coding 浓度
func processRecords(records [][]string) ([][]string, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	// 获取 GitHub token（可选）
	githubToken := os.Getenv("GH_TOKEN")
	if githubToken == "" {
		githubToken = os.Getenv("GITHUB_TOKEN")
	}

	// 创建 GitHub client
	client := github.NewClient()
	ctx := context.Background()

	// 构建输出记录（包含表头）
	result := [][]string{
		{"议题名称", "仓库地址", "合并PR总数", "AI生成PR数", "AI Coding浓度"},
	}

	// 跳过表头，处理数据行
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 2 {
			continue
		}

		topicName := record[0]
		repoURL := record[1]

		// 解析仓库地址
		repo := parseRepoURL(repoURL)
		if repo == "" {
			fmt.Fprintf(os.Stderr, "Warning: Invalid repo URL: %s\n", repoURL)
			result = append(result, []string{topicName, repoURL, "0", "0", "0.00%"})
			continue
		}

		// 获取所有 PR
		prs, _, err := client.GetPRList(ctx, repo, githubToken)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to get PRs for %s: %v\n", repo, err)
			result = append(result, []string{topicName, repoURL, "0", "0", "0.00%"})
			continue
		}

		// 统计合并的 PR
		totalMerged := 0
		botMerged := 0

		for _, pr := range prs {
			if pr.State == "merged" {
				totalMerged++
				if pr.Author.Login == "xgopilot[bot]" {
					botMerged++
				}
			}
		}

		// 计算百分比
		percentage := "0.00%"
		if totalMerged > 0 {
			pct := float64(botMerged) / float64(totalMerged) * 100
			percentage = fmt.Sprintf("%.2f%%", pct)
		}

		result = append(result, []string{
			topicName,
			repoURL,
			fmt.Sprintf("%d", totalMerged),
			fmt.Sprintf("%d", botMerged),
			percentage,
		})
	}

	return result, nil
}

// parseRepoURL 从 URL 中提取 owner/repo 格式
func parseRepoURL(url string) string {
	// 移除 .git 后缀
	url = strings.TrimSuffix(url, ".git")

	// 移除 https:// 和 http://
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	// 移除 github.com/
	url = strings.TrimPrefix(url, "github.com/")

	// 只取前两段 (owner/repo)
	parts := strings.Split(url, "/")
	if len(parts) >= 2 {
		return parts[0] + "/" + parts[1]
	}

	return ""
}
