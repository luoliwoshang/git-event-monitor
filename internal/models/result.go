package models

// AnalysisResult 分析结果
type AnalysisResult struct {
	Found            bool           `json:"found"`
	EventsChecked    int            `json:"events_checked"`
	LastCodeEvent    *UnifiedEvent  `json:"last_code_event,omitempty"`
	SubmittedBefore  *bool          `json:"submitted_before,omitempty"`
	TimeDifference   string         `json:"time_difference,omitempty"`
	EventDescription string         `json:"event_description,omitempty"`
	Error            string         `json:"error,omitempty"`
}

// Platform 平台类型
type Platform string

const (
	PlatformGitHub Platform = "github"
	PlatformGitee  Platform = "gitee"
)

// AnalysisRequest 分析请求
type AnalysisRequest struct {
	Repository string    `json:"repository"`
	Platform   Platform  `json:"platform"`
	Token      string    `json:"token,omitempty"`
	Deadline   string    `json:"deadline,omitempty"` // ISO 8601 格式
}