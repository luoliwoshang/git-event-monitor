import { MonitorRequest, CodeEventResult, PushEvent } from '../types'
import { formatDateTime } from '../utils/timeUtils'

// Gitee API 事件类型定义
interface GiteeEvent {
  id: string
  type: string
  actor: {
    id: number
    login: string
    display_name: string
    avatar_url: string
    url: string
    html_url: string
    followers_url: string
    following_url: string
    gists_url: string
    starred_url: string
    subscriptions_url: string
    organizations_url: string
    repos_url: string
    events_url: string
    received_events_url: string
    type: string
  }
  repo: {
    id: number
    name: string
    full_name: string
    owner: {
      id: number
      login: string
      name: string
      avatar_url: string
      url: string
      html_url: string
      remark: string
      followers_url: string
      following_url: string
      gists_url: string
      starred_url: string
      subscriptions_url: string
      organizations_url: string
      repos_url: string
      events_url: string
      received_events_url: string
      type: string
    }
    private: boolean
    html_url: string
    description: string
    fork: boolean
    url: string
  }
  payload: any
  public: boolean
  created_at: string
}

export class GiteeApi {
  private baseUrl = 'https://gitee.com/api/v5'

  async getRepositoryEvents(repository: string, token?: string): Promise<GiteeEvent[]> {
    const [owner, repo] = repository.split('/')
    if (!owner || !repo) {
      throw new Error('仓库名格式应为: owner/repo')
    }

    const url = new URL(`${this.baseUrl}/networks/${owner}/${repo}/events`)

    // Gitee 使用 access_token 参数而不是 Authorization header
    if (token) {
      url.searchParams.append('access_token', token)
    }

    const response = await fetch(url.toString())

    if (!response.ok) {
      if (response.status === 403) {
        throw new Error('API 请求被限制。请提供 Gitee 访问令牌以提高请求限制。')
      }
      const errorText = await response.text()
      throw new Error(`Gitee API 错误: ${response.status} - ${errorText}`)
    }

    const events: GiteeEvent[] = await response.json()
    return events.slice(0, 30) // 限制为最近30个事件
  }

  private isCodeSubmissionEvent(event: GiteeEvent): boolean {
    switch (event.type) {
      case 'PushEvent':
        // 所有 push 都算代码提交
        return true

      case 'PullRequestEvent':
        // 只有已合并的 PR 才算代码提交
        return event.payload.action === 'closed' &&
               event.payload.pull_request?.merged === true

      default:
        return false
    }
  }

  private convertToGitHubFormat(giteeEvent: GiteeEvent): PushEvent {
    // 将 Gitee 事件转换为与 GitHub 兼容的格式
    return {
      id: giteeEvent.id,
      type: giteeEvent.type as 'PushEvent',
      actor: {
        id: giteeEvent.actor.id,
        login: giteeEvent.actor.login,
        display_login: giteeEvent.actor.display_name || giteeEvent.actor.login,
        gravatar_id: '',
        url: giteeEvent.actor.url,
        avatar_url: giteeEvent.actor.avatar_url
      },
      repo: {
        id: giteeEvent.repo.id,
        name: giteeEvent.repo.full_name,
        url: giteeEvent.repo.url
      },
      payload: giteeEvent.payload,
      public: giteeEvent.public,
      created_at: giteeEvent.created_at
    }
  }

  async analyzeRepositoryCodeEvents(
    request: MonitorRequest
  ): Promise<CodeEventResult> {
    try {
      const events = await this.getRepositoryEvents(request.repository, request.token)
      console.log('Gitee events received:', events.length, events.map(e => ({ type: e.type, created_at: e.created_at })))

      // Filter for code submission events (push, merged PR)
      const codeEvents = events.filter(event => this.isCodeSubmissionEvent(event))
      console.log('Gitee code events filtered:', codeEvents.length, codeEvents.map(e => ({ type: e.type, created_at: e.created_at })))

      if (codeEvents.length === 0) {
        return {
          found: false,
          eventsChecked: events.length,
          error: `在最近的 ${events.length} 个仓库事件中未找到代码提交事件`
        }
      }

      const lastCodeEvent = codeEvents[0] // Most recent code event
      const result: CodeEventResult = {
        found: true,
        lastCodeEvent: this.convertToGitHubFormat(lastCodeEvent),
        eventsChecked: events.length,
        eventDescription: lastCodeEvent.type === 'PushEvent'
          ? `最近的推送事件 (${formatDateTime(lastCodeEvent.created_at)})`
          : `最近的合并事件 (${formatDateTime(lastCodeEvent.created_at)})`
      }

      // Check against deadline if provided
      if (request.deadline) {
        const eventTime = new Date(lastCodeEvent.created_at).getTime()
        const deadlineTime = request.deadline.getTime()
        const isBeforeDeadline = eventTime <= deadlineTime

        result.submittedBefore = isBeforeDeadline

        const timeDiff = Math.abs(eventTime - deadlineTime)
        const days = Math.floor(timeDiff / (1000 * 60 * 60 * 24))
        const hours = Math.floor((timeDiff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))
        const minutes = Math.floor((timeDiff % (1000 * 60 * 60)) / (1000 * 60))

        if (days > 0) {
          result.timeDifference = `${days}天${hours}小时`
        } else if (hours > 0) {
          result.timeDifference = `${hours}小时${minutes}分钟`
        } else {
          result.timeDifference = `${minutes}分钟`
        }

        if (isBeforeDeadline) {
          result.timeDifference = `截止时间前 ${result.timeDifference}`
        } else {
          result.timeDifference = `超过截止时间 ${result.timeDifference}`
        }
      }

      return result
    } catch (error) {
      return {
        found: false,
        eventsChecked: 0,
        error: error instanceof Error ? error.message : '未知错误'
      }
    }
  }
}