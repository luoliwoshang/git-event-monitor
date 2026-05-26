package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/luoliwoshang/git-event-monitor/internal/api/github"
)

func main() {
	// 定义命令行参数
	userDaily := flag.Bool("user-daily", false, "按用户按日统计 PR 数量")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <input.csv>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment variables:\n")
		fmt.Fprintf(os.Stderr, "  GH_TOKEN or GITHUB_TOKEN: GitHub access token (optional)\n")
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	inputFile := flag.Arg(0)

	// 读取输入 CSV
	records, err := readCSV(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read input CSV: %v\n", err)
		os.Exit(1)
	}

	// 根据模式处理记录
	var processedRecords [][]string
	if *userDaily {
		processedRecords, err = processUserDailyRecords(records)
	} else {
		processedRecords, err = processRecords(records)
	}
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
		{"议题名称", "仓库地址", "合并PR总数", "AI生成PR数", "AI Coding PR浓度"},
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
			result = append(result, []string{topicName, repoURL, "", "", ""})
			continue
		}

		// 获取所有 PR
		prs, _, err := client.GetPRList(ctx, repo, githubToken)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to get PRs for %s: %v\n", repo, err)
			result = append(result, []string{topicName, repoURL, "", "", ""})
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

// userDailyKey 用于按用户按日分组的 key
type userDailyKey struct {
	repoURL string
	user    string
	date    string
}

// userDailyStats 按用户按日的统计数据
type userDailyStats struct {
	createdCount    int
	mergedCount     int
	earliestCreated string // 最早 PR 创建时间 (ISO 8601)
	latestCreated   string // 最晚 PR 创建时间 (ISO 8601)
}

// processUserDailyRecords 处理输入记录，按用户按日统计 PR 数量
func processUserDailyRecords(records [][]string) ([][]string, error) {
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

	// 统计数据：repo -> user -> date -> stats
	statsMap := make(map[userDailyKey]*userDailyStats)

	// 跳过表头，处理数据行
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 2 {
			continue
		}

		repoURL := record[1]

		// 解析仓库地址
		repo := parseRepoURL(repoURL)
		if repo == "" {
			fmt.Fprintf(os.Stderr, "Warning: Invalid repo URL: %s\n", repoURL)
			continue
		}

		// 获取所有 PR
		prs, _, err := client.GetPRList(ctx, repo, githubToken)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to get PRs for %s: %v\n", repo, err)
			continue
		}

		// 遍历 PR，按用户按日统计
		for _, pr := range prs {
			user := pr.Author.Login
			if user == "" {
				user = "unknown"
			}

			// 统计创建的 PR（按 CreatedAt 日期）
			if pr.CreatedAt != "" {
				createdDate := extractDate(pr.CreatedAt)
				key := userDailyKey{repoURL: repoURL, user: user, date: createdDate}
				if statsMap[key] == nil {
					statsMap[key] = &userDailyStats{}
				}
				statsMap[key].createdCount++

				// 更新最早和最晚创建时间
				if statsMap[key].earliestCreated == "" || pr.CreatedAt < statsMap[key].earliestCreated {
					statsMap[key].earliestCreated = pr.CreatedAt
				}
				if statsMap[key].latestCreated == "" || pr.CreatedAt > statsMap[key].latestCreated {
					statsMap[key].latestCreated = pr.CreatedAt
				}
			}

			// 统计合并的 PR（按 MergedAt 日期）
			if pr.MergedAt != "" {
				mergedDate := extractDate(pr.MergedAt)
				key := userDailyKey{repoURL: repoURL, user: user, date: mergedDate}
				if statsMap[key] == nil {
					statsMap[key] = &userDailyStats{}
				}
				statsMap[key].mergedCount++
			}
		}
	}

	// 将 map 转换为有序的结果列表
	var keys []userDailyKey
	for key := range statsMap {
		keys = append(keys, key)
	}

	// 排序：按仓库 -> 用户 -> 日期
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].repoURL != keys[j].repoURL {
			return keys[i].repoURL < keys[j].repoURL
		}
		if keys[i].user != keys[j].user {
			return keys[i].user < keys[j].user
		}
		return keys[i].date < keys[j].date
	})

	// 构建输出记录（包含表头）
	result := [][]string{
		{"仓库", "用户", "日期", "创建PR数", "合并PR数", "最早PR创建时间", "最晚PR创建时间"},
	}

	prevRepo := ""
	for _, key := range keys {
		// 不同仓库之间插入空行
		if prevRepo != "" && prevRepo != key.repoURL {
			result = append(result, []string{"", "", "", "", "", "", ""})
		}
		prevRepo = key.repoURL

		stats := statsMap[key]
		result = append(result, []string{
			key.repoURL,
			key.user,
			key.date,
			fmt.Sprintf("%d", stats.createdCount),
			fmt.Sprintf("%d", stats.mergedCount),
			toChineseTime(stats.earliestCreated),
			toChineseTime(stats.latestCreated),
		})
	}

	return result, nil
}

// extractDate 从 ISO 8601 时间字符串中提取中国时间的日期部分 (YYYY-MM-DD)
func extractDate(isoTime string) string {
	if isoTime == "" {
		return ""
	}

	// 解析 ISO 8601 时间
	t, err := time.Parse(time.RFC3339, isoTime)
	if err != nil {
		// 解析失败时，尝试直接截取前10位
		if len(isoTime) >= 10 {
			return isoTime[:10]
		}
		return isoTime
	}

	// 转换为中国时区 (UTC+8)
	chinaLoc := time.FixedZone("CST", 8*60*60)
	chinaTime := t.In(chinaLoc)

	// 返回中国时间的日期部分
	return chinaTime.Format("2006-01-02")
}

// toChineseTime 将 ISO 8601 时间转换为中国时间 (UTC+8)
func toChineseTime(isoTime string) string {
	if isoTime == "" {
		return ""
	}

	// 解析 ISO 8601 时间
	t, err := time.Parse(time.RFC3339, isoTime)
	if err != nil {
		return isoTime
	}

	// 转换为中国时区 (UTC+8)
	chinaLoc := time.FixedZone("CST", 8*60*60)
	chinaTime := t.In(chinaLoc)

	// 格式化为易读格式：YYYY-MM-DD HH:MM:SS
	return chinaTime.Format("2006-01-02 15:04:05")
}
