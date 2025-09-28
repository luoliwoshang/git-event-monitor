import {
  Box,
  Container,
  Heading,
  VStack,
  Card,
  CardBody,
  Text,
  useColorModeValue,
} from '@chakra-ui/react'
import { RepositoryForm } from './RepositoryForm'
import { ResultDisplay } from './ResultDisplay'
import { useState } from 'react'
import { PushEventResult } from '../types'

export function GitEventMonitor() {
  const [result, setResult] = useState<PushEventResult | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  const bgColor = useColorModeValue('gray.50', 'gray.900')
  const cardBg = useColorModeValue('white', 'gray.800')

  const handleAnalyze = (result: PushEventResult) => {
    setResult(result)
    setIsLoading(false)
  }

  const handleLoading = (loading: boolean) => {
    setIsLoading(loading)
  }

  return (
    <Box minH="100vh" bg={bgColor} py={8}>
      <Container maxW="6xl">
        <VStack spacing={8} align="stretch">
          {/* Header */}
          <Box textAlign="center">
            <Heading
              as="h1"
              size="2xl"
              bgGradient="linear(to-r, brand.500, brand.600)"
              bgClip="text"
              mb={4}
            >
              Git Event Monitor
            </Heading>
            <Text fontSize="lg" color="gray.600">
              Monitor Git repository push events for code competition fairness
            </Text>
          </Box>

          {/* Main Content */}
          <Card bg={cardBg} shadow="lg">
            <CardBody>
              <VStack spacing={6} align="stretch">
                <RepositoryForm
                  onAnalyze={handleAnalyze}
                  onLoading={handleLoading}
                  isLoading={isLoading}
                />

                {result && (
                  <ResultDisplay result={result} />
                )}
              </VStack>
            </CardBody>
          </Card>

          {/* Footer */}
          <Text textAlign="center" color="gray.500" fontSize="sm">
            Powered by GitHub API • Built with React + Chakra UI
          </Text>
        </VStack>
      </Container>
    </Box>
  )
}