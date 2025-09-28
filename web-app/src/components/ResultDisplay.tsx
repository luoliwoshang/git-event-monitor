import { PushEventResult } from '../types'
import { formatDateTime } from '../utils/timeUtils'
import './ResultDisplay.css'

interface ResultDisplayProps {
  result: PushEventResult
}

export function ResultDisplay({ result }: ResultDisplayProps) {
  if (!result.found) {
    return (
      <div className="result-display">
        <div className="warning-alert">
          <span className="icon">⚠️</span>
          <div className="content">
            <div className="title">No Push Events Found</div>
            <div className="message">{result.error}</div>
            <div className="stats">Checked {result.eventsChecked} events</div>
          </div>
        </div>
      </div>
    )
  }

  const { lastPushEvent, eventsChecked, pushedBefore, timeDifference } = result

  return (
    <div className="result-display">
      {/* Summary */}
      <div className="summary-card">
        <div className="summary-header">
          <span className="summary-title">Analysis Summary</span>
          <span className="events-badge">{eventsChecked} events checked</span>
        </div>

        {pushedBefore !== undefined && (
          <div className="deadline-status">
            <span className={`status-badge ${pushedBefore ? 'success' : 'error'}`}>
              {pushedBefore ? '✓ Before Deadline' : '✗ After Deadline'}
            </span>
            {timeDifference && <span className="time-diff">{timeDifference}</span>}
          </div>
        )}
      </div>

      {lastPushEvent && (
        <>
          <div className="divider"></div>

          {/* Push Event Details */}
          <div className="details-section">
            <h3 className="details-title">Last Push Event Details</h3>

            <div className="details-card">
              {/* Time and Actor */}
              <div className="detail-row">
                <div className="detail-item">
                  <div className="detail-label">🕐 Push Time</div>
                  <code className="detail-value">{formatDateTime(lastPushEvent.created_at)}</code>
                </div>
                <div className="detail-item">
                  <div className="detail-label">Pushed by</div>
                  <a
                    href={`https://github.com/${lastPushEvent.actor.login}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="user-link"
                  >
                    {lastPushEvent.actor.login} ↗
                  </a>
                </div>
              </div>

              {/* Branch */}
              {lastPushEvent.payload.ref && (
                <div className="detail-item">
                  <div className="detail-label">🌿 Branch</div>
                  <code className="detail-value">{lastPushEvent.payload.ref.replace('refs/heads/', '')}</code>
                </div>
              )}

              {/* Commits */}
              {lastPushEvent.payload.commits && lastPushEvent.payload.commits.length > 0 && (
                <div className="commits-section">
                  <div className="detail-label">
                    📝 Commits ({lastPushEvent.payload.size || lastPushEvent.payload.commits.length})
                  </div>
                  <div className="commits-list">
                    {lastPushEvent.payload.commits.slice(0, 3).map((commit) => (
                      <div key={commit.sha} className="commit-item">
                        <div className="commit-header">
                          <code className="commit-sha">{commit.sha.substring(0, 7)}</code>
                          <span className="commit-author">{commit.author.name}</span>
                        </div>
                        <div className="commit-message">{commit.message}</div>
                      </div>
                    ))}
                    {lastPushEvent.payload.commits.length > 3 && (
                      <div className="more-commits">
                        ... and {lastPushEvent.payload.commits.length - 3} more commits
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
          </div>
        </>
      )}
    </div>
  )
}