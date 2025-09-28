# Git 仓库最后提交时间检查工具设计文档

## 需求概述
通过 GitHub/Gitee API 检查某个仓库最后代码提交的真实时间，不依赖可被修改的 commit 时间戳。

## 核心需求
1. 获取仓库最后代码推送事件的真实时间（代码活跃事件）
2. 不能基于 commit 时间来判断（因为 commit 时间可以被修改）
3. 同时支持 GitHub 和 Gitee 平台
4. 提供命令行界面
5. 检查所有分支，确保没有任何分支在截止时间后有更新

## 使用场景
代码比赛公平性监控：确保在截止时间（DDL）后没有代码更新

## 技术方案
### API 接口选择
- **GitHub**: `/repos/{owner}/{repo}/events` - 获取仓库事件（包括 PushEvent）
  - 返回最近 30 天内最多 300 个事件
  - 支持 `per_page` 参数（最大100，默认30），建议使用 per_page=100
  - PushEvent 包含推送时间、分支信息等
  - 文档：https://docs.github.com/en/rest/activity/events#list-repository-events
  - 事件类型：https://docs.github.com/en/rest/using-the-rest-api/github-event-types?apiVersion=2022-11-28
- **Gitee**: `/api/v5/repos/{owner}/{repo}/events` - 列出仓库的所有公开动态
  - 需要 access token 认证
  - 支持 `limit` 参数，默认使用 limit=100
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

### 工具类型
命令行工具

## 实现目标
获取指定仓库最后一次代码推送到远程仓库的真实时间