import {
  Box,
  VStack,
  HStack,
  Text,
  Badge,
  Divider,
  Alert,
  AlertIcon,
  Code,
  Link,
  Icon,
  useColorModeValue,
} from '@chakra-ui/react'
import { ExternalLinkIcon, CalendarIcon, GitBranch } from '@chakra-ui/icons'
import { PushEventResult } from '../types'
import { formatDateTime } from '../utils/timeUtils'

interface ResultDisplayProps {
  result: PushEventResult
}

export function ResultDisplay({ result }: ResultDisplayProps) {
  const cardBg = useColorModeValue('gray.50', 'gray.700')
  const borderColor = useColorModeValue('gray.200', 'gray.600')

  if (!result.found) {
    return (
      <Alert status="warning">
        <AlertIcon />
        <VStack align="start" spacing={2}>
          <Text fontWeight="bold">No Push Events Found</Text>
          <Text>{result.error}</Text>
          <Text fontSize="sm">Checked {result.eventsChecked} events</Text>
        </VStack>
      </Alert>
    )
  }

  const { lastPushEvent, eventsChecked, pushedBefore, timeDifference } = result

  return (
    <VStack spacing={6} align="stretch">
      {/* Summary */}
      <Box p={4} bg={cardBg} borderRadius="md" borderWidth={1} borderColor={borderColor}>
        <VStack spacing={3} align="start">
          <HStack>
            <Text fontWeight="bold">Analysis Summary</Text>
            <Badge colorScheme="blue">{eventsChecked} events checked</Badge>
          </HStack>

          {pushedBefore !== undefined && (
            <HStack>
              <Badge colorScheme={pushedBefore ? 'green' : 'red'} size="lg">
                {pushedBefore ? '✓ Before Deadline' : '✗ After Deadline'}
              </Badge>
              {timeDifference && <Text>{timeDifference}</Text>}
            </HStack>
          )}
        </VStack>
      </Box>

      {lastPushEvent && (
        <>
          <Divider />

          {/* Push Event Details */}
          <VStack spacing={4} align="stretch">
            <Text fontSize="lg" fontWeight="bold">Last Push Event Details</Text>

            <Box p={4} bg={cardBg} borderRadius="md" borderWidth={1} borderColor={borderColor}>
              <VStack spacing={3} align="stretch">
                {/* Time and Actor */}
                <HStack justify="space-between">
                  <HStack>
                    <Icon as={CalendarIcon} />
                    <VStack align="start" spacing={0}>
                      <Text fontWeight="bold">Push Time</Text>
                      <Code>{formatDateTime(lastPushEvent.created_at)}</Code>
                    </VStack>
                  </HStack>
                  <VStack align="end" spacing={0}>
                    <Text fontSize="sm" color="gray.600">Pushed by</Text>
                    <Link
                      href={`https://github.com/${lastPushEvent.actor.login}`}
                      isExternal
                      color="brand.500"
                    >
                      {lastPushEvent.actor.login} <ExternalLinkIcon mx="2px" />
                    </Link>
                  </VStack>
                </HStack>

                {/* Branch */}
                {lastPushEvent.payload.ref && (
                  <HStack>
                    <Icon as={GitBranch} />
                    <VStack align="start" spacing={0}>
                      <Text fontWeight="bold">Branch</Text>
                      <Code>{lastPushEvent.payload.ref.replace('refs/heads/', '')}</Code>
                    </VStack>
                  </HStack>
                )}

                {/* Commits */}
                {lastPushEvent.payload.commits && lastPushEvent.payload.commits.length > 0 && (
                  <VStack align="stretch" spacing={2}>
                    <Text fontWeight="bold">
                      Commits ({lastPushEvent.payload.size || lastPushEvent.payload.commits.length})
                    </Text>
                    {lastPushEvent.payload.commits.slice(0, 3).map((commit) => (
                      <Box
                        key={commit.sha}
                        p={3}
                        bg={useColorModeValue('white', 'gray.800')}
                        borderRadius="md"
                        borderWidth={1}
                        borderColor={borderColor}
                      >
                        <VStack align="stretch" spacing={1}>
                          <HStack justify="space-between">
                            <Code fontSize="xs">{commit.sha.substring(0, 7)}</Code>
                            <Text fontSize="xs" color="gray.600">{commit.author.name}</Text>
                          </HStack>
                          <Text fontSize="sm">{commit.message}</Text>
                        </VStack>
                      </Box>
                    ))}
                    {lastPushEvent.payload.commits.length > 3 && (
                      <Text fontSize="sm" color="gray.600" textAlign="center">
                        ... and {lastPushEvent.payload.commits.length - 3} more commits
                      </Text>
                    )}
                  </VStack>
                )}
              </VStack>
            </Box>
          </VStack>
        </>
      )}
    </VStack>
  )
}