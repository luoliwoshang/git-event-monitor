package models

// Commit 代表一个 Git 提交
// 目前为空结构体，后续可以扩展添加字段（如 SHA, Author, Message, Date 等）
type Commit struct{}

// PRState 表示 Pull Request 的状态
type PRState string

const (
	// PRStateOpen PR 处于打开状态
	PRStateOpen PRState = "open"
	// PRStateClosed PR 已关闭但未合并
	PRStateClosed PRState = "closed"
	// PRStateMerged PR 已合并
	PRStateMerged PRState = "merged"
)

// PullRequest 代表一个 Pull Request / Merge Request
type PullRequest struct {
	// State PR 状态：open, closed, merged
	State PRState
	// Author PR 作者信息
	Author Author
}

// Author 代表作者信息
type Author struct {
	// Login 用户名
	Login string
	// Name 显示名称（可能为空）
	Name string
	// Email 邮箱地址（可能为空）
	Email string
}

// PRStats PR 统计信息
type PRStats struct {
	Total  int // 总数
	Open   int // 打开的 PR 数量
	Closed int // 关闭但未合并的 PR 数量
	Merged int // 已合并的 PR 数量
}
