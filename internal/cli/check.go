package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/luoliwoshang/git-event-monitor/internal/api"
	"github.com/luoliwoshang/git-event-monitor/internal/api/github"
	"github.com/luoliwoshang/git-event-monitor/internal/api/gitee"
	"github.com/luoliwoshang/git-event-monitor/internal/models"
	"github.com/luoliwoshang/git-event-monitor/internal/output"
)

var (
	platform string
	token    string
	deadline string
	format   string
)

var checkCmd = &cobra.Command{
	Use:   "check <owner/repo>",
	Short: "Check repository code submission events",
	Long: `Check the latest code submission events for a repository.

Examples:
  git-event-monitor check microsoft/vscode
  git-event-monitor check microsoft/vscode --platform github --token ghp_xxxxx
  git-event-monitor check owner/repo --platform gitee --deadline "2024-03-15T18:00:00Z"`,
	Args: cobra.ExactArgs(1),
	RunE: runCheck,
}

func init() {
	checkCmd.Flags().StringVar(&platform, "platform", "github", "Platform to check (github or gitee)")
	checkCmd.Flags().StringVar(&token, "token", "", "API token (optional for public repos)")
	checkCmd.Flags().StringVar(&deadline, "deadline", "", "Deadline for compliance check (ISO 8601 format)")
	checkCmd.Flags().StringVar(&format, "output", "table", "Output format (table or json)")
}

func runCheck(cmd *cobra.Command, args []string) error {
	repo := args[0]

	// 验证仓库名格式
	if !strings.Contains(repo, "/") {
		return fmt.Errorf("repository format should be 'owner/repo'")
	}

	// 验证平台
	var platformType models.Platform
	switch platform {
	case "github":
		platformType = models.PlatformGitHub
	case "gitee":
		platformType = models.PlatformGitee
	default:
		return fmt.Errorf("unsupported platform: %s (supported: github, gitee)", platform)
	}

	// 创建分析请求
	req := &models.AnalysisRequest{
		Repository: repo,
		Platform:   platformType,
		Token:      token,
		Deadline:   deadline,
	}

	// 创建对应平台的客户端
	var client api.Client
	switch platformType {
	case models.PlatformGitHub:
		client = github.NewClient()
	case models.PlatformGitee:
		client = gitee.NewClient()
	}

	// 执行分析
	ctx := context.Background()
	result, err := client.AnalyzeCodeEvents(ctx, req)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// 输出结果
	formatter := output.NewFormatter(format)
	return formatter.Format(result)
}