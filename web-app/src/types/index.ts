export interface GitHubEvent {
  id: string
  type: string
  actor: {
    id: number
    login: string
    display_login: string
    gravatar_id: string
    url: string
    avatar_url: string
  }
  repo: {
    id: number
    name: string
    url: string
  }
  payload: {
    push_id?: number
    size?: number
    distinct_size?: number
    ref?: string
    head?: string
    before?: string
    commits?: Array<{
      sha: string
      author: {
        email: string
        name: string
      }
      message: string
      distinct: boolean
      url: string
    }>
  }
  public: boolean
  created_at: string
}

// 重命名为更通用的类型，支持多种代码提交事件
export type PushEvent = GitHubEvent

export interface CodeEventResult {
  found: boolean
  eventsChecked: number
  lastCodeEvent?: PushEvent
  submittedBefore?: boolean
  timeDifference?: string
  eventDescription?: string
  error?: string
}

// 保持向后兼容
export type PushEventResult = CodeEventResult

export interface MonitorRequest {
  repository: string
  platform: 'github' | 'gitee'
  token?: string
  deadline?: Date
}