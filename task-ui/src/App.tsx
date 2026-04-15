import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { TaskListPage } from './pages/TaskListPage'
import './App.css'

const queryClient = new QueryClient()

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <TaskListPage agencyId="f4511042-c4e8-4baa-a27f-a4c84aea0976" />
    </QueryClientProvider>
  )
}

export default App
