import { Card, Alert, Tag, Descriptions, Timeline, Typography, Space, Row, Col } from 'antd'
import { CheckCircleOutlined, CloseCircleOutlined, ClockCircleOutlined, UserOutlined, BranchesOutlined, FileTextOutlined, ExclamationCircleOutlined } from '@ant-design/icons'
import { CodeEventResult } from '../types'
import { formatDateTime, formatLocalDateTime } from '../utils/timeUtils'

const { Text, Link } = Typography

interface ResultDisplayProps {
  result: CodeEventResult
}

export function ResultDisplay({ result }: ResultDisplayProps) {
  if (!result.found) {
    return (
      <Card
        style={{
          boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)',
          borderRadius: '8px'
        }}
      >
        <Alert
          message="No Code Events Found"
          description={
            <div>
              <Text>{result.error}</Text>
              <br />
              <Text type="secondary">Checked {result.eventsChecked} events</Text>
            </div>
          }
          type="warning"
          icon={<ExclamationCircleOutlined />}
          showIcon
        />
      </Card>
    )
  }

  const { lastCodeEvent, eventsChecked, submittedBefore, timeDifference, eventDescription } = result

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      {/* Analysis Summary */}
      <Card
        title="Analysis Summary"
        extra={<Tag color="blue">{eventsChecked} events checked</Tag>}
        style={{
          boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)',
          borderRadius: '8px'
        }}
      >
        {submittedBefore !== undefined && (
          <Row gutter={[16, 16]} align="middle">
            <Col xs={24} sm={12}>
              <Tag
                icon={submittedBefore ? <CheckCircleOutlined /> : <CloseCircleOutlined />}
                color={submittedBefore ? 'success' : 'error'}
                style={{
                  fontSize: 'clamp(12px, 2.5vw, 14px)',
                  padding: '8px 12px'
                }}
              >
                {submittedBefore ? 'Before Deadline' : 'After Deadline'}
              </Tag>
            </Col>
            {timeDifference && (
              <Col xs={24} sm={12}>
                <Text type="secondary" style={{ fontSize: 'clamp(12px, 2.5vw, 14px)' }}>
                  <ClockCircleOutlined /> {timeDifference}
                </Text>
              </Col>
            )}
          </Row>
        )}

        {eventDescription && (
          <Row>
            <Col span={24}>
              <Text strong style={{ fontSize: 'clamp(14px, 3vw, 16px)' }}>
                {eventDescription}
              </Text>
            </Col>
          </Row>
        )}
      </Card>

      {/* Code Event Details */}
      {lastCodeEvent && (
        <Card
          title="Last Code Event Details"
          style={{
            boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)',
            borderRadius: '8px'
          }}
        >
          <Descriptions bordered column={{ xs: 1, sm: 1, md: 1 }} size="small">
            <Descriptions.Item
              label={
                <Space>
                  <ClockCircleOutlined />
                  Event Time
                </Space>
              }
            >
              <Space direction="vertical" size="small">
                <div>
                  <Text type="secondary" style={{ fontSize: '12px' }}>GitHub 原始时间：</Text>
                  <Text code style={{ marginLeft: '8px' }}>{lastCodeEvent.created_at}</Text>
                </div>
                <div>
                  <Text type="secondary" style={{ fontSize: '12px' }}>UTC 时间：</Text>
                  <Text code style={{ marginLeft: '8px' }}>{formatDateTime(lastCodeEvent.created_at)}</Text>
                </div>
                <div>
                  <Text type="secondary" style={{ fontSize: '12px' }}>本地时间：</Text>
                  <Text code style={{ marginLeft: '8px' }}>{formatLocalDateTime(lastCodeEvent.created_at)}</Text>
                </div>
              </Space>
            </Descriptions.Item>

            <Descriptions.Item
              label={
                <Space>
                  <UserOutlined />
                  Author
                </Space>
              }
            >
              <Link
                href={`https://github.com/${lastCodeEvent.actor.login}`}
                target="_blank"
              >
                {lastCodeEvent.actor.login}
              </Link>
            </Descriptions.Item>

            {lastCodeEvent.payload.ref && (
              <Descriptions.Item
                label={
                  <Space>
                    <BranchesOutlined />
                    Branch
                  </Space>
                }
              >
                <Tag color="geekblue">
                  {lastCodeEvent.payload.ref.replace('refs/heads/', '')}
                </Tag>
              </Descriptions.Item>
            )}

            {lastCodeEvent.payload.commits && lastCodeEvent.payload.commits.length > 0 && (
              <Descriptions.Item
                label={
                  <Space>
                    <FileTextOutlined />
                    Commits ({lastCodeEvent.payload.size || lastCodeEvent.payload.commits.length})
                  </Space>
                }
                span={3}
              >
                <Timeline
                  items={[
                    ...lastCodeEvent.payload.commits.slice(0, 3).map((commit) => ({
                      children: (
                        <div key={commit.sha}>
                          <Space direction="vertical" size={4}>
                            <Space>
                              <Text code style={{ fontSize: '12px' }}>
                                {commit.sha.substring(0, 7)}
                              </Text>
                              <Text type="secondary" style={{ fontSize: '12px' }}>
                                {commit.author.name}
                              </Text>
                            </Space>
                            <Text>{commit.message}</Text>
                          </Space>
                        </div>
                      )
                    })),
                    ...(lastCodeEvent.payload.commits.length > 3 ? [{
                      children: (
                        <Text type="secondary">
                          ... and {lastCodeEvent.payload.commits.length - 3} more commits
                        </Text>
                      )
                    }] : [])
                  ]}
                />
              </Descriptions.Item>
            )}
          </Descriptions>
        </Card>
      )}
    </Space>
  )
}