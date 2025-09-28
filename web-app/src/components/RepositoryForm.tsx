import { useState } from 'react'
import { Card, Form, Input, Button, Select, DatePicker, Space, Typography, Row, Col, Alert } from 'antd'
import { SearchOutlined, CalendarOutlined, LinkOutlined } from '@ant-design/icons'
import { GitHubApiService } from '../services/githubApi'
import { GiteeApi } from '../services/giteeApi'
import { MonitorRequest, CodeEventResult } from '../types'
import dayjs from 'dayjs'

const { Text } = Typography
const { Option } = Select

interface RepositoryFormProps {
  onAnalyze: (result: CodeEventResult) => void
  onLoading: (loading: boolean) => void
  loading: boolean
}

export function RepositoryForm({ onAnalyze, onLoading, loading }: RepositoryFormProps) {
  const [form] = Form.useForm()
  const [platform, setPlatform] = useState<'github' | 'gitee'>('github')

  const handleSubmit = async (values: any) => {
    onLoading(true)

    try {
      const request: MonitorRequest = {
        repository: values.repository,
        platform: platform,
        token: values.token,
        deadline: values.deadline ? values.deadline.toDate() : undefined
      }

      let result: CodeEventResult
      if (platform === 'github') {
        const githubApi = new GitHubApiService()
        result = await githubApi.analyzeRepositoryCodeEvents(request)
      } else {
        const giteeApi = new GiteeApi()
        result = await giteeApi.analyzeRepositoryCodeEvents(request)
      }

      onAnalyze(result)
    } catch (error) {
      onAnalyze({
        found: false,
        eventsChecked: 0,
        error: error instanceof Error ? error.message : 'Unknown error occurred'
      })
    } finally {
      onLoading(false)
    }
  }

  const getTokenHelpText = () => {
    if (platform === 'github') {
      return (
        <Space direction="vertical" size="small">
          <Text type="secondary" style={{ fontSize: '12px' }}>
            推荐用于公开仓库，私有仓库必需。
          </Text>
          <Text>
            <LinkOutlined /> <a href="https://github.com/settings/tokens" target="_blank" rel="noopener noreferrer">
              获取令牌 →
            </a>
          </Text>
        </Space>
      )
    } else {
      return (
        <Space direction="vertical" size="small">
          <Text type="secondary" style={{ fontSize: '12px' }}>
            推荐用于公开仓库，私有仓库必需。
          </Text>
          <Text>
            <LinkOutlined /> <a href="https://gitee.com/personal_access_tokens" target="_blank" rel="noopener noreferrer">
              获取令牌 →
            </a>
          </Text>
        </Space>
      )
    }
  }

  return (
    <Card
      title="Repository Analysis"
      style={{
        boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)',
        borderRadius: '8px',
        minWidth: '400px'
      }}
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
        size="large"
      >
        <Form.Item
          label="Repository"
          name="repository"
          rules={[{ required: true, message: 'Please enter repository name' }]}
        >
          <Input
            placeholder="owner/repository"
            prefix={<SearchOutlined />}
            style={{ fontSize: 'clamp(14px, 2.5vw, 16px)' }}
          />
        </Form.Item>

        <Form.Item label="Platform">
          <Select
            value={platform}
            onChange={(value) => setPlatform(value)}
            style={{ fontSize: 'clamp(14px, 2.5vw, 16px)' }}
          >
            <Option value="github">GitHub</Option>
            <Option value="gitee">Gitee</Option>
          </Select>
        </Form.Item>

        <Form.Item
          label="API Token (Recommended)"
          name="token"
          help={getTokenHelpText()}
        >
          <Input.Password
            placeholder="ghp_xxxxx or access_token"
            style={{ fontSize: 'clamp(14px, 2.5vw, 16px)' }}
          />
        </Form.Item>

        <Form.Item
          label="Set Deadline for Compliance Check"
          name="deadline"
        >
          <DatePicker
            showTime
            placeholder="Select deadline"
            prefix={<CalendarOutlined />}
            style={{
              width: '100%',
              fontSize: 'clamp(14px, 2.5vw, 16px)'
            }}
            disabledDate={(current) => current && current < dayjs().subtract(1, 'year')}
          />
        </Form.Item>

        <Form.Item>
          <Row>
            <Col span={24}>
              <Button
                type="primary"
                htmlType="submit"
                loading={loading}
                block
                size="large"
                style={{
                  background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                  border: 'none',
                  fontSize: 'clamp(14px, 2.5vw, 16px)',
                  height: '48px'
                }}
              >
                {loading ? 'Analyzing...' : 'Analyze Repository'}
              </Button>
            </Col>
          </Row>
        </Form.Item>
      </Form>

      <Alert
        message="How it works"
        description="This tool checks repository events to monitor code submissions. It analyzes push events and merged pull requests to determine compliance with deadlines."
        type="info"
        showIcon
        style={{ marginTop: '16px' }}
      />
    </Card>
  )
}