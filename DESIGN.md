# Git 仓库最后提交时间检查工具设计文档

## 需求概述
通过 GitHub/Gitee API 检查某个仓库最后代码提交的真实时间，不依赖可被修改的 commit 时间戳。

## 核心需求
1. 获取仓库最后代码推送事件的真实时间（代码活跃事件）
2. 不能基于 commit 时间来判断（因为 commit 时间可以被修改）
3. 同时支持 GitHub 和 Gitee 平台
4. 提供 Web 界面，通过 GitHub Pages 部署
5. 检查所有分支，确保没有任何分支在截止时间后有更新

## 使用场景
代码比赛公平性监控：确保在截止时间（DDL）后没有代码更新

---

# CLI 工具设计（Go 版本）

## 项目概述

基于现有 Web 版本功能，开发一个功能更强大的 Go 命令行工具，用于监控 Git 仓库的代码提交事件，支持 GitHub 和 Gitee 平台。

## 架构设计

### 项目结构
```
git-event-monitor/
├── cmd/
│   └── git-event-monitor/
│       └── main.go           # 程序入口，初始化 cobra 根命令
└── internal/                 # 所有核心逻辑（不对外暴露）
    ├── models/               # 统一数据模型
    │   ├── event.go         # 统一事件模型
    │   ├── repository.go    # 仓库信息模型
    │   ├── actor.go         # 操作者模型
    │   └── result.go        # 分析结果模型
    ├── api/                  # API 客户端层
    │   ├── github/          # GitHub API 客户端 + 数据转换
    │   ├── gitee/           # Gitee API 客户端 + 数据转换
    │   └── client.go        # 通用 API 接口定义
    ├── monitor/              # 事件监控和分析逻辑
    │   ├── analyzer.go      # 事件分析器
    │   └── filter.go        # 事件过滤器
    ├── config/               # 配置管理
    │   └── config.go        # CLI 参数和配置处理
    ├── output/               # 输出格式化
    │   ├── formatter.go     # 输出格式化接口
    │   ├── table.go         # 表格输出
    │   └── json.go          # JSON 输出
    └── cli/                  # CLI 命令层
        ├── root.go          # 根命令定义
        └── check.go         # check 命令实现
```

## 核心功能设计

### 第一阶段：基础检查功能
复制 Web 版本的核心能力，提供单个仓库的事件检查功能。

#### CLI 命令接口
```bash
# 基础检查
git-event-monitor check owner/repo

# 指定平台（默认 github）
git-event-monitor check owner/repo --platform github
git-event-monitor check owner/repo --platform gitee

# 带截止时间检查
git-event-monitor check owner/repo --deadline "2024-03-15T18:00:00Z"

# 使用 API token
git-event-monitor check owner/repo --token ghp_xxxxx

# 指定输出格式
git-event-monitor check owner/repo --output table  # 默认
git-event-monitor check owner/repo --output json
```

## 数据模型设计

### 统一事件模型
基于 GitHub 和 Gitee API 的共同字段，设计统一的事件数据结构：

```go
// Event 统一事件模型
type Event struct {
    ID          string    `json:"id"`
    Type        EventType `json:"type"`
    CreatedAt   time.Time `json:"created_at"`
    Actor       Actor     `json:"actor"`
    Repository  Repository `json:"repository"`
    Payload     EventPayload `json:"payload"`
}

// EventType 事件类型
type EventType string

const (
    EventTypePush        EventType = "PushEvent"
    EventTypePullRequest EventType = "PullRequestEvent"
    // 其他需要的事件类型...
)

// Actor 操作者信息
type Actor struct {
    ID       int64  `json:"id"`
    Login    string `json:"login"`
    Name     string `json:"name,omitempty"`
    AvatarURL string `json:"avatar_url,omitempty"`
}

// Repository 仓库信息
type Repository struct {
    ID       int64  `json:"id"`
    Name     string `json:"name"`
    FullName string `json:"full_name"`
    URL      string `json:"url,omitempty"`
}

// EventPayload 事件载荷（不同事件类型有不同结构）
type EventPayload interface {
    GetEventType() EventType
}

// PushEventPayload Push 事件载荷
type PushEventPayload struct {
    Ref     string   `json:"ref"`
    Size    int      `json:"size"`
    Commits []Commit `json:"commits"`
}

// Commit 提交信息
type Commit struct {
    SHA     string `json:"sha"`
    Message string `json:"message"`
    Author  Author `json:"author"`
}

// Author 提交作者
type Author struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}
```

### 分析结果模型
```go
// AnalysisResult 分析结果
type AnalysisResult struct {
    Found           bool      `json:"found"`
    EventsChecked   int       `json:"events_checked"`
    LastCodeEvent   *Event    `json:"last_code_event,omitempty"`
    SubmittedBefore *bool     `json:"submitted_before,omitempty"`
    TimeDifference  string    `json:"time_difference,omitempty"`
    EventDescription string   `json:"event_description,omitempty"`
    Error           string    `json:"error,omitempty"`
}
```

## API 层设计

### 通用 API 接口
```go
// Client API 客户端通用接口
type Client interface {
    GetRepositoryEvents(ctx context.Context, repo string, opts *RequestOptions) ([]models.Event, error)
    AnalyzeCodeEvents(ctx context.Context, req *AnalysisRequest) (*models.AnalysisResult, error)
}

// RequestOptions API 请求选项
type RequestOptions struct {
    Token    string
    PerPage  int
    Platform Platform
}

// AnalysisRequest 分析请求
type AnalysisRequest struct {
    Repository string
    Platform   Platform
    Token      string
    Deadline   *time.Time
}

// Platform 平台类型
type Platform string

const (
    PlatformGitHub Platform = "github"
    PlatformGitee  Platform = "gitee"
)
```

### 平台特定实现
每个平台的 API 客户端负责：
1. 调用平台特定的 API
2. 将平台返回的数据转换为统一的 `models.Event` 结构
3. 处理平台特定的错误和限制

## 技术选型

- **CLI 框架**: `github.com/spf13/cobra`
- **配置管理**: `github.com/spf13/viper`
- **HTTP 客户端**: 标准库 `net/http`
- **JSON 处理**: 标准库 `encoding/json`
- **表格输出**: `github.com/olekukonko/tablewriter`
- **时间处理**: 标准库 `time`

## 开发阶段规划

### 阶段 1: 核心功能实现
- [ ] 搭建项目基础架构
- [ ] 实现统一数据模型
- [ ] 实现 GitHub API 客户端
- [ ] 实现 Gitee API 客户端
- [ ] 实现事件分析逻辑
- [ ] 实现 CLI check 命令
- [ ] 实现基础输出格式

### 阶段 2: 功能增强（未来扩展）
- [ ] 批量检查多个仓库
- [ ] 配置文件支持
- [ ] 缓存机制
- [ ] 更多输出格式
- [ ] 持续监控功能

## 设计原则

1. **单一职责**: 每个模块专注于特定功能
2. **接口抽象**: 通过接口隔离平台差异
3. **数据统一**: 使用统一模型简化上层逻辑
4. **可扩展性**: 便于添加新平台和新功能
5. **测试友好**: 依赖注入，便于单元测试

---

# Web 应用设计（React 版本）

## 技术方案
### API 接口选择
- **GitHub**: `/repos/{owner}/{repo}/events` - 获取仓库事件（包括 PushEvent）
  - 返回最近 30 天内最多 300 个事件
  - 支持 `per_page` 参数（最大100，默认30），建议使用 per_page=100
  - 支持 CORS，可从浏览器直接调用
  - Token 可选（公开仓库不需要，私有仓库或提高频率限制需要）
  - PushEvent 包含推送时间、分支信息等
  - 文档：https://docs.github.com/en/rest/activity/events#list-repository-events
  - 事件类型：https://docs.github.com/en/rest/using-the-rest-api/github-event-types?apiVersion=2022-11-28
- **Gitee**: `/api/v5/repos/{owner}/{repo}/events` - 列出仓库的所有公开动态
  - Token 可选（公开仓库不需要，私有仓库需要）
  - 支持 `limit` 参数，默认使用 limit=100
  - 可能存在 CORS 限制，需要在实现时验证
  - 数据结构与 GitHub 类似，包含 `created_at`、`type`、`payload.ref` 等字段
  - WebHook 文档：https://help.gitee.com/webhook/gitee-webhook-push-data-format/

### 数据来源
使用 API 返回的推送事件时间戳，而非 git commit 中的时间戳
- **获取方式**:
  - 两个平台都不支持服务器端事件类型过滤
  - 需要获取所有事件然后客户端过滤 `type: "PushEvent"`
  - **错误处理**: 如果在检查范围内（GitHub per_page=100，Gitee limit=100）没有找到 PushEvent，提示错误
- 关键字段：`created_at` - 事件被触发的时间（服务器端时间戳，无法伪造）
  - GitHub: ISO 8601 格式，如 `"2022-06-09T12:47:28Z"`
  - Gitee: 包含时区信息，如 `"2023-07-09T03:00:43+08:00"`
- 事件类型：过滤 `type: "PushEvent"`
- 分支信息：`payload.ref` 包含分支信息

### 应用类型
React Web 应用，部署在 GitHub Pages

### 输入界面设计
- 仓库路径输入框（格式: owner/repo）
- 平台选择（GitHub | Gitee）
- API Token 输入框（可选，用于私有仓库或提高频率限制）
- 截止时间输入框（可选，ISO 8601 格式，如 `2024-03-15T18:00:00Z`）

### 输出内容设计
1. **检查统计**: 检查了多少个事件（limit 数量）
2. **时间判断**: 最后一个 PushEvent 是否在 deadline 之前
3. **时间差异**: 在 deadline 之前/之后多长时间
4. **提交信息**: 最后推送的 commit 详细信息
5. **分支信息**: 哪个分支有最后的推送

### 技术选择
- **前端框架**: React + TypeScript
- **UI 库**: Ant Design（现代化设计）
- **HTTP 请求**: Fetch API 或 Axios
- **时间处理**: Day.js
- **部署**: GitHub Pages
- **构建工具**: Vite

## 实现目标
获取指定仓库最后一次代码推送到远程仓库的真实时间