import axios from 'axios'
import { GitHubEvent, CodeEventResult, MonitorRequest } from '../types'
import { formatDateTime } from '../utils/timeUtils'

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
        if (error.response?.status === 403) {
          throw new Error('API rate limit exceeded. Please provide a GitHub token to increase the rate limit.')
        }
        throw new Error(`GitHub API Error: ${error.response?.status} - ${error.response?.statusText}`)
      }
      throw new Error('Failed to fetch repository events')
    }
  }

  private isCodeSubmissionEvent(event: GitHubEvent): boolean {
    switch (event.type) {
      case 'PushEvent':
        return true
      case 'PullRequestEvent':
        // Only count merged PRs as code submissions
        return (event.payload as any).action === 'closed' &&
               (event.payload as any).pull_request?.merged === true
      default:
        return false
    }
  }

  async analyzeRepositoryCodeEvents(
    request: MonitorRequest
  ): Promise<CodeEventResult> {
    try {
      const events = await this.getRepositoryEvents(request.repository, request.token)

      // Filter for code submission events (push, merged PR)
      const codeEvents = events.filter(event => this.isCodeSubmissionEvent(event))

      if (codeEvents.length === 0) {
        return {
          found: false,
          eventsChecked: events.length,
          error: `No code submission events found in the last ${events.length} repository events`
        }
      }

      const lastCodeEvent = codeEvents[0] // Most recent code event
      const result: CodeEventResult = {
        found: true,
        eventsChecked: events.length,
        lastCodeEvent,
        eventDescription: lastCodeEvent.type === 'PushEvent'
          ? `Latest push event (${formatDateTime(lastCodeEvent.created_at)})`
          : `Latest merge event (${formatDateTime(lastCodeEvent.created_at)})`
      }

      // If deadline is provided, check if submission was before deadline
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
          result.timeDifference = `${days} days ${hours} hours`
        } else if (hours > 0) {
          result.timeDifference = `${hours} hours ${minutes} minutes`
        } else {
          result.timeDifference = `${minutes} minutes`
        }

        if (isBeforeDeadline) {
          result.timeDifference = `${result.timeDifference} before deadline`
        } else {
          result.timeDifference = `${result.timeDifference} after deadline`
        }
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