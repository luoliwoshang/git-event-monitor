import { useState } from 'react'
import { MonitorRequest } from '../types'
import { GitHubApiService } from '../services/githubApi'
import { isValidDateTime } from '../utils/timeUtils'
import './RepositoryForm.css'

interface RepositoryFormProps {
  onAnalyze: (result: any) => void
  onLoading: (loading: boolean) => void
  isLoading: boolean
}

export function RepositoryForm({ onAnalyze, onLoading, isLoading }: RepositoryFormProps) {
  const [repository, setRepository] = useState('')
  const [platform, setPlatform] = useState<'github' | 'gitee'>('github')
  const [token, setToken] = useState('')
  const [deadline, setDeadline] = useState('')
  const [error, setError] = useState('')

  const githubApi = new GitHubApiService()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    // Validation
    if (!repository.trim()) {
      setError('Repository is required')
      return
    }

    if (!repository.includes('/')) {
      setError('Repository must be in format: owner/repo')
      return
    }

    if (deadline && !isValidDateTime(deadline)) {
      setError('Invalid deadline format. Use ISO 8601 format like: 2024-03-15T18:00:00Z')
      return
    }

    const request: MonitorRequest = {
      repository: repository.trim(),
      platform,
      token: token.trim() || undefined,
      deadline: deadline.trim() || undefined,
    }

    onLoading(true)

    try {
      const result = await githubApi.analyzeRepositoryPushEvents(request)
      onAnalyze(result)
    } catch (error) {
      setError(error instanceof Error ? error.message : 'An error occurred')
    } finally {
      onLoading(false)
    }
  }

  return (
    <div className="repository-form">
      <form onSubmit={handleSubmit}>
        {error && (
          <div className="error-alert">
            ⚠️ {error}
          </div>
        )}

        <div className="form-group">
          <label htmlFor="repository">Repository *</label>
          <input
            id="repository"
            type="text"
            placeholder="owner/repository"
            value={repository}
            onChange={(e) => setRepository(e.target.value)}
            required
          />
          <small>Enter repository in format: owner/repo</small>
        </div>

        <div className="form-group">
          <label htmlFor="platform">Platform</label>
          <select
            id="platform"
            value={platform}
            onChange={(e) => setPlatform(e.target.value as 'github' | 'gitee')}
          >
            <option value="github">GitHub</option>
            <option value="gitee" disabled>Gitee (Coming Soon)</option>
          </select>
        </div>

        <div className="form-group">
          <label htmlFor="token">API Token (Optional)</label>
          <input
            id="token"
            type="password"
            placeholder="ghp_xxxxxxxxxxxx"
            value={token}
            onChange={(e) => setToken(e.target.value)}
          />
          <small>Optional. Increases rate limits and allows private repos</small>
        </div>

        <div className="form-group">
          <label htmlFor="deadline">Deadline (Optional)</label>
          <input
            id="deadline"
            type="text"
            placeholder="2024-03-15T18:00:00Z"
            value={deadline}
            onChange={(e) => setDeadline(e.target.value)}
          />
          <small>ISO 8601 format. Leave empty to just show last push time</small>
        </div>

        <button
          type="submit"
          className={`submit-button ${isLoading ? 'loading' : ''}`}
          disabled={isLoading}
        >
          {isLoading ? 'Analyzing...' : 'Analyze Repository'}
        </button>
      </form>
    </div>
  )
}