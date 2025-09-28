import { RepositoryForm } from './RepositoryForm'
import { ResultDisplay } from './ResultDisplay'
import { useState } from 'react'
import { PushEventResult } from '../types'
import './GitEventMonitor.css'

export function GitEventMonitor() {
  const [result, setResult] = useState<PushEventResult | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  const handleAnalyze = (result: PushEventResult) => {
    setResult(result)
    setIsLoading(false)
  }

  const handleLoading = (loading: boolean) => {
    setIsLoading(loading)
  }

  return (
    <div className="git-event-monitor">
      <div className="container">
        {/* Header */}
        <div className="header">
          <h1 className="title">Git Event Monitor</h1>
          <p className="subtitle">
            Monitor Git repository push events for code competition fairness
          </p>
        </div>

        {/* Main Content */}
        <div className="main-content">
          <RepositoryForm
            onAnalyze={handleAnalyze}
            onLoading={handleLoading}
            isLoading={isLoading}
          />

          {result && (
            <div className="result-section">
              <ResultDisplay result={result} />
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="footer">
          Powered by GitHub API • Built with React + Vite
        </div>
      </div>
    </div>
  )
}