import { RepositoryForm } from './components/RepositoryForm'
import { ResultDisplay } from './components/ResultDisplay'
import { useState } from 'react'
import { CodeEventResult } from './types'
import { Row, Col, Layout, Typography } from 'antd'
import './index.css'

const { Header, Content } = Layout
const { Title } = Typography

function App() {
  const [result, setResult] = useState<CodeEventResult | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  const handleAnalyze = (analysisResult: CodeEventResult) => {
    setResult(analysisResult)
  }

  const handleLoading = (loading: boolean) => {
    setIsLoading(loading)
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        padding: '0 24px',
        display: 'flex',
        alignItems: 'center'
      }}>
        <Title level={2} style={{ color: 'white', margin: 0 }}>
          Git Event Monitor
        </Title>
      </Header>

      <Content style={{
        padding: '24px',
        background: 'linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%)',
        minHeight: 'calc(100vh - 64px)'
      }}>
        <Row gutter={[24, 24]} style={{ height: '100%' }}>
          <Col xs={24} lg={12} xl={10}>
            <RepositoryForm
              onAnalyze={handleAnalyze}
              onLoading={handleLoading}
              loading={isLoading}
            />
          </Col>
          <Col xs={24} lg={12} xl={14}>
            {result && <ResultDisplay result={result} />}
          </Col>
        </Row>
      </Content>
    </Layout>
  )
}

export default App