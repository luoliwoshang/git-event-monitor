package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/luoliwoshang/git-event-monitor/internal/api"
	"github.com/luoliwoshang/git-event-monitor/internal/api/gitee"
	"github.com/luoliwoshang/git-event-monitor/internal/api/github"
	"github.com/luoliwoshang/git-event-monitor/internal/models"
	"github.com/xuri/excelize/v2"
)

// 使用示例:
// go run cmd/csv-processor/main.go --gitee-token=YOUR_GITEE_TOKEN --github-token=YOUR_GITHUB_TOKEN --deadline=2025-09-30T23:59:59Z "议题三 待筛选 名单.xlsx" 2 67

func main() {
	// 定义命令行参数
	var githubToken = flag.String("github-token", "", "GitHub API token")
	var giteeToken = flag.String("gitee-token", "", "Gitee API token")
	var deadline = flag.String("deadline", "", "Deadline in RFC3339 format (e.g., 2024-03-15T18:00:00Z)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <csv-file> <start-row> <end-row>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nArguments:\n")
		fmt.Fprintf(os.Stderr, "  csv-file    Path to the CSV file\n")
		fmt.Fprintf(os.Stderr, "  start-row   Starting row number (1-indexed, >=2)\n")
		fmt.Fprintf(os.Stderr, "  end-row     Ending row number (1-indexed, inclusive)\n")
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s --github-token=ghp_xxx --gitee-token=xxx --deadline=2024-03-15T18:00:00Z data.csv 2 4\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nNote: start-row and end-row are 1-indexed (header is row 1, first data is row 2)\n")
	}

	flag.Parse()

	// 检查必需的位置参数
	if flag.NArg() < 3 {
		flag.Usage()
		os.Exit(1)
	}

	filename := flag.Arg(0)
	startRow := parseInt(flag.Arg(1))
	endRow := parseInt(flag.Arg(2))

	// 验证行号参数
	if startRow < 2 {
		fmt.Println("❌ Start row must be >= 2 (row 1 is header)")
		os.Exit(1)
	}
	if endRow < startRow {
		fmt.Println("❌ End row must be >= start row")
		os.Exit(1)
	}

	fmt.Printf("🚀 Starting CSV processing...\n")
	fmt.Printf("File: %s\n", filename)
	fmt.Printf("Processing rows: %d to %d\n", startRow, endRow)
	if *deadline != "" {
		fmt.Printf("Deadline: %s\n", *deadline)
	}
	if *githubToken != "" {
		fmt.Printf("GitHub Token: %s\n", maskToken(*githubToken))
	}
	if *giteeToken != "" {
		fmt.Printf("Gitee Token: %s\n", maskToken(*giteeToken))
	}
	fmt.Println()

	// 读取文件（支持CSV和Excel）
	records, err := readFile(filename)
	if err != nil {
		fmt.Printf("❌ 文件读取失败: %v\n", err)
		os.Exit(1)
	}

	if len(records) < 2 {
		fmt.Println("❌ 文件至少需要包含表头和一行数据")
		return
	}

	// 找到相关列的索引
	headers := records[0]
	repoColumnIndex := findColumnIndex(headers, "代码仓库地址")
	nameColumnIndex := findColumnIndex(headers, "姓名")

	// 寻找可访问性和提交状态列，如果不存在则添加
	accessColumnIndex := findColumnIndex(headers, "是否可访问")
	submissionColumnIndex := findColumnIndex(headers, "是否准时提交")

	if repoColumnIndex == -1 {
		fmt.Println("❌ 未找到'代码仓库地址'列")
		return
	}

	// 如果结果列不存在，添加到表头
	if accessColumnIndex == -1 {
		headers = append(headers, "是否可访问")
		accessColumnIndex = len(headers) - 1
		// 为所有现有行添加空列
		for i := 1; i < len(records); i++ {
			records[i] = append(records[i], "")
		}
		fmt.Printf("📝 添加新列: 是否可访问 (第%d列)\n", accessColumnIndex+1)
	}

	if submissionColumnIndex == -1 {
		headers = append(headers, "是否准时提交")
		submissionColumnIndex = len(headers) - 1
		// 为所有现有行添加空列
		for i := 1; i < len(records); i++ {
			records[i] = append(records[i], "")
		}
		fmt.Printf("📝 添加新列: 是否准时提交 (第%d列)\n", submissionColumnIndex+1)
	}

	// 更新表头
	records[0] = headers

	fmt.Printf("📍 列位置:\n")
	fmt.Printf("  代码仓库地址: 第%d列\n", repoColumnIndex+1)
	fmt.Printf("  是否可访问: 第%d列\n", accessColumnIndex+1)
	fmt.Printf("  是否准时提交: 第%d列\n", submissionColumnIndex+1)
	if nameColumnIndex != -1 {
		fmt.Printf("  姓名: 第%d列\n", nameColumnIndex+1)
	}
	fmt.Println()

	// 验证行号范围是否有效
	if endRow > len(records) {
		fmt.Printf("❌ End row %d exceeds total rows %d\n", endRow, len(records))
		os.Exit(1)
	}

	// 处理指定范围的记录 (转换为0-based索引)
	startIndex := startRow - 1 // 转换为0-based索引
	endIndex := endRow         // endRow本身就是包含的，所以不需要-1

	fmt.Printf("📊 Processing %d records (data rows %d to %d)...\n\n", endIndex-startIndex, startRow, endRow)

	for i := startIndex; i < endIndex; i++ {
		record := records[i]
		if len(record) <= repoColumnIndex {
			continue
		}

		name := ""
		if nameColumnIndex != -1 && len(record) > nameColumnIndex {
			name = record[nameColumnIndex]
		}

		repoURL := record[repoColumnIndex]
		fmt.Printf("📦 Processing row %d: %s\n", i+1, name)
		fmt.Printf("   Repository: %s\n", repoURL)

		// 解析仓库 URL
		platform, owner, repo := parseRepositoryURL(repoURL)
		if platform == "" {
			// 对于无法解析的URL（多个URL、非GitHub/Gitee、格式错误等），
			// 只输出日志，不更新CSV行
			fmt.Printf("   ⏭️  Skipping: Cannot parse repository URL (multiple URLs, unsupported platform, or invalid format)\n")
			fmt.Println()
			continue
		}

		repoPath := fmt.Sprintf("%s/%s", owner, repo)
		fmt.Printf("   Platform: %s, Repository: %s\n", platform, repoPath)

		// 创建对应平台的客户端并选择对应Token
		var client api.Client
		var platformType models.Platform
		var currentToken string
		switch platform {
		case "github":
			client = github.NewClient()
			platformType = models.PlatformGitHub
			currentToken = *githubToken
		case "gitee":
			client = gitee.NewClient()
			platformType = models.PlatformGitee
			currentToken = *giteeToken
		default:
			// 理论上不会到这里，因为parseRepositoryURL已经过滤了不支持的平台
			// 如果到了这里，说明代码有bug，只输出日志不更新CSV
			fmt.Printf("   ❌ Internal error: Unsupported platform: %s\n", platform)
			fmt.Println()
			continue
		}

		// 检查是否可访问
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_, err := client.GetEvents(ctx, repoPath, currentToken)
		cancel()

		if err != nil {
			fmt.Printf("   ❌ Repository not accessible: %v\n", err)
			updateRecord(record, accessColumnIndex, "不可访问")
			// 不可访问时，准时提交列留空，不做任何更新
			fmt.Println()
			continue
		}

		fmt.Printf("   ✅ Repository accessible\n")
		updateRecord(record, accessColumnIndex, "可访问")

		// 如果没有截止时间，跳过提交时间检查
		if *deadline == "" {
			fmt.Printf("   ⏭️  No deadline specified, skipping submission check\n")
			updateRecord(record, submissionColumnIndex, "未设置截止时间")
			fmt.Println()
			continue
		}

		// 检查是否准时提交
		req := &models.AnalysisRequest{
			Repository: repoPath,
			Platform:   platformType,
			Token:      currentToken,
			Deadline:   *deadline,
		}

		ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
		result, err := client.AnalyzeCodeEvents(ctx2, req)
		cancel2()

		if err != nil {
			fmt.Printf("   ❌ Analysis failed: %v\n", err)
			updateRecord(record, submissionColumnIndex, "分析失败")
		} else if !result.Found {
			// 没有找到PushEvent，需要进一步检查仓库是否有提交记录
			fmt.Printf("   ⚠️  No push events found in recent activity\n")

			// 调用HasCommits API检查仓库是否有提交记录
			ctx3, cancel3 := context.WithTimeout(context.Background(), 10*time.Second)
			hasCommits, commitErr := client.HasCommits(ctx3, repoPath, currentToken)
			cancel3()

			if commitErr != nil {
				// HasCommits API调用失败，记录为分析失败
				fmt.Printf("   ❌ Failed to check commits: %v\n", commitErr)
				updateRecord(record, submissionColumnIndex, "分析失败")
			} else if hasCommits {
				// 仓库有提交记录但没有PushEvent，可能是初始提交或批量提交
				fmt.Printf("   ℹ️  Repository has commits but no recent push events (likely initial commit)\n")
				updateRecord(record, submissionColumnIndex, "初始提交（无法检查提交时间）")
			} else {
				// 仓库没有任何提交记录，是空仓库
				fmt.Printf("   ⚠️  Repository is empty (no commits found)\n")
				updateRecord(record, submissionColumnIndex, "空仓库（无法检查提交时间）")
			}
		} else if result.SubmittedBefore == nil {
			fmt.Printf("   ⚠️  Could not determine submission time\n")
			updateRecord(record, submissionColumnIndex, "无法确定")
		} else if *result.SubmittedBefore {
			fmt.Printf("   ✅ Submitted before deadline (%s)\n", result.TimeDifference)
			updateRecord(record, submissionColumnIndex, "准时提交")
		} else {
			fmt.Printf("   ❌ Submitted after deadline (%s)\n", result.TimeDifference)
			updateRecord(record, submissionColumnIndex, "超时提交")
		}

		fmt.Println()
	}

	// 写入更新后的文件
	err = writeFile(filename, records)
	if err != nil {
		fmt.Printf("❌ 文件写入失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 处理完成！结果已保存\n")
}

// parseRepositoryURL 解析仓库 URL，返回平台、owner、repo
// 支持 GitHub 和 Gitee 的多种格式：
// GitHub HTTPS: https://github.com/owner/repo 或 https://github.com/owner/repo.git
// GitHub SSH: git@github.com:owner/repo.git
// GitHub Plain: github.com/owner/repo 或 github.com/owner/repo.git
// Gitee HTTPS: https://gitee.com/owner/repo 或 https://gitee.com/owner/repo.git
// Gitee SSH: git@gitee.com:owner/repo.git
// Gitee Plain: gitee.com/owner/repo 或 gitee.com/owner/repo.git
// 对于多个URL、非标准格式、不支持的平台等情况返回空字符串
func parseRepositoryURL(url string) (platform, owner, repo string) {
	// 清理 URL，去除首尾空格
	url = strings.TrimSpace(url)

	// 检查是否包含多个URL（通过换行符、空格、多个http等判断）
	if strings.Contains(url, "\n") || strings.Count(url, "http") > 1 {
		// 包含多个URL，直接跳过
		return "", "", ""
	}

	// 检查URL长度是否合理（避免处理过长或过短的无效输入）
	if len(url) < 10 || len(url) > 200 {
		return "", "", ""
	}

	// GitHub URL 模式匹配
	// 支持格式：
	// HTTPS: https://github.com/owner/repo 或 https://github.com/owner/repo.git
	// SSH: git@github.com:owner/repo.git
	// Plain: github.com/owner/repo 或 github.com/owner/repo.git
	githubHTTPSPattern := regexp.MustCompile(`(?i)^https?://github\.com[/:]([^/\s]+)/([^/\s]+?)(?:\.git)?/?$`)
	githubSSHPattern := regexp.MustCompile(`(?i)^git@github\.com:([^/\s]+)/([^/\s]+?)(?:\.git)?/?$`)
	githubPlainPattern := regexp.MustCompile(`(?i)^github\.com[/:]([^/\s]+)/([^/\s]+?)(?:\.git)?/?$`)

	if matches := githubHTTPSPattern.FindStringSubmatch(url); len(matches) == 3 {
		owner := strings.TrimSpace(matches[1])
		repo := strings.TrimSpace(matches[2])
		// 验证owner和repo名称的有效性（不能为空，不能包含特殊字符）
		if owner != "" && repo != "" && isValidRepoName(owner) && isValidRepoName(repo) {
			return "github", owner, repo
		}
	}

	if matches := githubSSHPattern.FindStringSubmatch(url); len(matches) == 3 {
		owner := strings.TrimSpace(matches[1])
		repo := strings.TrimSpace(matches[2])
		// 验证owner和repo名称的有效性
		if owner != "" && repo != "" && isValidRepoName(owner) && isValidRepoName(repo) {
			return "github", owner, repo
		}
	}

	if matches := githubPlainPattern.FindStringSubmatch(url); len(matches) == 3 {
		owner := strings.TrimSpace(matches[1])
		repo := strings.TrimSpace(matches[2])
		// 验证owner和repo名称的有效性
		if owner != "" && repo != "" && isValidRepoName(owner) && isValidRepoName(repo) {
			return "github", owner, repo
		}
	}

	// Gitee URL 模式匹配
	// 支持格式：
	// HTTPS: https://gitee.com/owner/repo 或 https://gitee.com/owner/repo.git
	// SSH: git@gitee.com:owner/repo.git
	// Plain: gitee.com/owner/repo 或 gitee.com/owner/repo.git
	giteeHTTPSPattern := regexp.MustCompile(`(?i)^https?://gitee\.com[/:]([^/\s]+)/([^/\s]+?)(?:\.git)?/?$`)
	giteeSSHPattern := regexp.MustCompile(`(?i)^git@gitee\.com:([^/\s]+)/([^/\s]+?)(?:\.git)?/?$`)
	giteePlainPattern := regexp.MustCompile(`(?i)^gitee\.com[/:]([^/\s]+)/([^/\s]+?)(?:\.git)?/?$`)

	if matches := giteeHTTPSPattern.FindStringSubmatch(url); len(matches) == 3 {
		owner := strings.TrimSpace(matches[1])
		repo := strings.TrimSpace(matches[2])
		// 验证owner和repo名称的有效性
		if owner != "" && repo != "" && isValidRepoName(owner) && isValidRepoName(repo) {
			return "gitee", owner, repo
		}
	}

	if matches := giteeSSHPattern.FindStringSubmatch(url); len(matches) == 3 {
		owner := strings.TrimSpace(matches[1])
		repo := strings.TrimSpace(matches[2])
		// 验证owner和repo名称的有效性
		if owner != "" && repo != "" && isValidRepoName(owner) && isValidRepoName(repo) {
			return "gitee", owner, repo
		}
	}

	if matches := giteePlainPattern.FindStringSubmatch(url); len(matches) == 3 {
		owner := strings.TrimSpace(matches[1])
		repo := strings.TrimSpace(matches[2])
		// 验证owner和repo名称的有效性
		if owner != "" && repo != "" && isValidRepoName(owner) && isValidRepoName(repo) {
			return "gitee", owner, repo
		}
	}

	// 无法解析或不支持的格式
	return "", "", ""
}

// isValidRepoName 验证仓库名称是否有效
// 仓库名称应该只包含字母、数字、连字符、下划线和点号
func isValidRepoName(name string) bool {
	if name == "" {
		return false
	}
	// 简单的仓库名称验证：只允许字母数字、连字符、下划线、点号
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	return validPattern.MatchString(name)
}

// findColumnIndex 查找列的索引
func findColumnIndex(headers []string, columnName string) int {
	for i, header := range headers {
		if strings.Contains(header, columnName) {
			return i
		}
	}
	return -1
}

// updateRecord 更新记录中的指定列
func updateRecord(record []string, columnIndex int, value string) {
	if columnIndex != -1 && columnIndex < len(record) {
		record[columnIndex] = value
	}
}

// parseInt 解析字符串为整数，出错时退出程序
func parseInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf("❌ Invalid number: %s\n", s)
		os.Exit(1)
	}
	return val
}

// maskToken 隐藏Token的敏感部分，只显示前几位和后几位
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}

// readFile 读取文件内容，支持CSV和Excel格式
// 返回二维字符串数组，第一行为表头，后续为数据行
func readFile(filename string) ([][]string, error) {
	// 根据文件扩展名判断文件类型
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".csv":
		return readCSVFile(filename)
	case ".xlsx", ".xls":
		return readExcelFile(filename)
	default:
		return nil, fmt.Errorf("不支持的文件格式: %s（支持.csv, .xlsx, .xls）", ext)
	}
}

// readCSVFile 读取CSV文件
func readCSVFile(filename string) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("无法打开CSV文件: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("读取CSV内容失败: %w", err)
	}

	return records, nil
}

// readExcelFile 读取Excel文件
func readExcelFile(filename string) ([][]string, error) {
	// 打开Excel文件
	file, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, fmt.Errorf("无法打开Excel文件: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("⚠️ 关闭Excel文件时出错: %v\n", err)
		}
	}()

	// 获取第一个工作表的所有行
	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel文件中没有工作表")
	}

	sheetName := sheets[0] // 使用第一个工作表
	fmt.Printf("📖 正在读取工作表: %s\n", sheetName)

	rows, err := file.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取工作表失败: %w", err)
	}

	return rows, nil
}

// writeFile 写入文件，根据原文件格式决定输出格式
func writeFile(originalFilename string, records [][]string) error {
	ext := strings.ToLower(filepath.Ext(originalFilename))

	switch ext {
	case ".csv":
		return writeCSVFile(originalFilename, records)
	case ".xlsx", ".xls":
		return writeExcelFile(originalFilename, records)
	default:
		return fmt.Errorf("不支持的文件格式: %s", ext)
	}
}

// writeCSVFile 写入CSV文件
func writeCSVFile(originalFilename string, records [][]string) error {
	outputFile := strings.TrimSuffix(originalFilename, filepath.Ext(originalFilename)) + "_processed.csv"
	fmt.Printf("💾 保存结果到: %s\n", outputFile)

	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("无法创建输出文件: %w", err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	err = writer.WriteAll(records)
	if err != nil {
		return fmt.Errorf("写入CSV失败: %w", err)
	}

	return nil
}

// writeExcelFile 写入Excel文件
func writeExcelFile(originalFilename string, records [][]string) error {
	outputFile := strings.TrimSuffix(originalFilename, filepath.Ext(originalFilename)) + "_processed.xlsx"
	fmt.Printf("💾 保存结果到: %s\n", outputFile)

	// 创建新的Excel文件
	file := excelize.NewFile()
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("⚠️ 关闭Excel文件时出错: %v\n", err)
		}
	}()

	sheetName := "Sheet1"

	// 写入所有行数据
	for rowIndex, row := range records {
		for colIndex, cellValue := range row {
			// Excel使用1-based索引
			cellName, err := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			if err != nil {
				return fmt.Errorf("生成单元格坐标失败: %w", err)
			}

			err = file.SetCellValue(sheetName, cellName, cellValue)
			if err != nil {
				return fmt.Errorf("设置单元格值失败: %w", err)
			}
		}
	}

	// 保存文件
	err := file.SaveAs(outputFile)
	if err != nil {
		return fmt.Errorf("保存Excel文件失败: %w", err)
	}

	return nil
}
