import { ChakraProvider } from '@chakra-ui/react'
import { GitEventMonitor } from './components/GitEventMonitor'
import theme from './theme'

function App() {
  return (
    <ChakraProvider theme={theme}>
      <GitEventMonitor />
    </ChakraProvider>
  )
}

export default App
