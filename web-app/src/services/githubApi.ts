import axios from 'axios'
import { GitHubEvent, PushEventResult, MonitorRequest } from '../types'
import { formatTimeDifference } from '../utils/timeUtils'

export class GitHubApiService {
  private baseUrl = 'https://api.github.com'

  async getRepositoryEvents(
    repository: string,
    token?: string
  ): Promise<GitHubEvent[]> {
    const headers: Record<string, string> = {
      'Accept': 'application/vnd.github.v3+json',
      'User-Agent': 'Git-Event-Monitor/1.0'
    }

    if (token) {
      headers['Authorization'] = `token ${token}`
    }

    try {
      const response = await axios.get(
        `${this.baseUrl}/repos/${repository}/events`,
        {
          headers,
          params: {
            per_page: 100
          }
        }
      )

      return response.data
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new Error(`GitHub API Error: ${error.response?.status} - ${error.response?.statusText}`)
      }
      throw new Error('Failed to fetch repository events')
    }
  }

  async analyzeRepositoryPushEvents(
    request: MonitorRequest
  ): Promise<PushEventResult> {
    try {
      const events = await this.getRepositoryEvents(request.repository, request.token)

      // Filter for push events
      const pushEvents = events.filter(event => event.type === 'PushEvent')

      if (pushEvents.length === 0) {
        return {
          found: false,
          eventsChecked: events.length,
          error: `No push events found in the last ${events.length} repository events`
        }
      }

      const lastPushEvent = pushEvents[0] // Most recent push event
      const result: PushEventResult = {
        found: true,
        eventsChecked: events.length,
        lastPushEvent
      }

      // If deadline is provided, check if push was before deadline
      if (request.deadline) {
        const pushTime = new Date(lastPushEvent.created_at)
        const deadlineTime = new Date(request.deadline)

        result.pushedBefore = pushTime < deadlineTime
        result.timeDifference = formatTimeDifference(pushTime, deadlineTime)
      }

      return result
    } catch (error) {
      return {
        found: false,
        eventsChecked: 0,
        error: error instanceof Error ? error.message : 'Unknown error occurred'
      }
    }
  }
}