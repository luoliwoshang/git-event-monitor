package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"

	"github.com/luoliwoshang/git-event-monitor/internal/models"
)

// Formatter 输出格式化器接口
type Formatter interface {
	Format(result *models.AnalysisResult) error
}

// NewFormatter 创建格式化器
func NewFormatter(format string) Formatter {
	switch format {
	case "json":
		return &JSONFormatter{}
	case "table":
		fallthrough
	default:
		return &TableFormatter{}
	}
}

// TableFormatter 表格格式化器
type TableFormatter struct{}

// Format 格式化为表格输出
func (t *TableFormatter) Format(result *models.AnalysisResult) error {
	if !result.Found {
		fmt.Printf("❌ No code events found\n")
		fmt.Printf("📊 Events checked: %d\n", result.EventsChecked)
		if result.Error != "" {
			fmt.Printf("❗ Error: %s\n", result.Error)
		}
		return nil
	}

	fmt.Printf("✅ Code event found\n")
	fmt.Printf("📊 Events checked: %d\n", result.EventsChecked)

	if result.EventDescription != "" {
		fmt.Printf("📝 %s\n", result.EventDescription)
	}

	if result.SubmittedBefore != nil {
		if *result.SubmittedBefore {
			fmt.Printf("⏰ Status: ✅ Submitted before deadline\n")
		} else {
			fmt.Printf("⏰ Status: ❌ Submitted after deadline\n")
		}
	}

	if result.TimeDifference != "" {
		fmt.Printf("📅 Time difference: %s\n", result.TimeDifference)
	}

	// 显示事件详情表格
	if result.LastCodeEvent != nil {
		fmt.Printf("\n📋 Last Code Event Details:\n")
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Field", "Value"})
		table.SetBorder(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)

		table.Append([]string{"Event ID", result.LastCodeEvent.ID})
		table.Append([]string{"Event Type", result.LastCodeEvent.Type})
		table.Append([]string{"Created At", result.LastCodeEvent.CreatedAt})
		table.Append([]string{"Actor", result.LastCodeEvent.ActorLogin})
		table.Append([]string{"Repository", result.LastCodeEvent.RepoName})

		table.Render()
	}

	return nil
}

// JSONFormatter JSON 格式化器
type JSONFormatter struct{}

// Format 格式化为 JSON 输出
func (j *JSONFormatter) Format(result *models.AnalysisResult) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}