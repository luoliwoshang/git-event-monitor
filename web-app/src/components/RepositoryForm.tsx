import {
  VStack,
  FormControl,
  FormLabel,
  Input,
  Select,
  Button,
  Alert,
  AlertIcon,
  FormHelperText,
} from '@chakra-ui/react'
import { useState } from 'react'
import { MonitorRequest } from '../types'
import { GitHubApiService } from '../services/githubApi'
import { isValidDateTime } from '../utils/timeUtils'

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
    <form onSubmit={handleSubmit}>
      <VStack spacing={4} align="stretch">
        {error && (
          <Alert status="error">
            <AlertIcon />
            {error}
          </Alert>
        )}

        <FormControl isRequired>
          <FormLabel>Repository</FormLabel>
          <Input
            placeholder="owner/repository"
            value={repository}
            onChange={(e) => setRepository(e.target.value)}
          />
          <FormHelperText>Enter repository in format: owner/repo</FormHelperText>
        </FormControl>

        <FormControl>
          <FormLabel>Platform</FormLabel>
          <Select
            value={platform}
            onChange={(e) => setPlatform(e.target.value as 'github' | 'gitee')}
          >
            <option value="github">GitHub</option>
            <option value="gitee" disabled>Gitee (Coming Soon)</option>
          </Select>
        </FormControl>

        <FormControl>
          <FormLabel>API Token (Optional)</FormLabel>
          <Input
            type="password"
            placeholder="ghp_xxxxxxxxxxxx"
            value={token}
            onChange={(e) => setToken(e.target.value)}
          />
          <FormHelperText>Optional. Increases rate limits and allows private repos</FormHelperText>
        </FormControl>

        <FormControl>
          <FormLabel>Deadline (Optional)</FormLabel>
          <Input
            placeholder="2024-03-15T18:00:00Z"
            value={deadline}
            onChange={(e) => setDeadline(e.target.value)}
          />
          <FormHelperText>ISO 8601 format. Leave empty to just show last push time</FormHelperText>
        </FormControl>

        <Button
          type="submit"
          colorScheme="brand"
          size="lg"
          isLoading={isLoading}
          loadingText="Analyzing..."
        >
          Analyze Repository
        </Button>
      </VStack>
    </form>
  )
}