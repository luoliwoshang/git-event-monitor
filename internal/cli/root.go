package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "git-event-monitor",
	Short: "Monitor Git repository code submission events",
	Long: `A tool to monitor Git repository code submission events for code competition fairness.

Supports both GitHub and Gitee platforms, checking push events and merged pull requests
to verify code submission compliance with deadlines.`,
}

// Execute 执行根命令
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// 添加子命令
	rootCmd.AddCommand(checkCmd)
}

// 如果命令执行出错，打印使用说明
func init() {
	cobra.OnInitialize()

	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})
}