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

export interface PushEventResult {
  found: boolean
  eventsChecked: number
  lastPushEvent?: GitHubEvent
  pushedBefore?: boolean
  timeDifference?: string
  error?: string
}

export interface MonitorRequest {
  repository: string
  platform: 'github' | 'gitee'
  token?: string
  deadline?: string
}