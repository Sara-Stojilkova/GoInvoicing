import { BrowserRouter, Routes, Route } from "react-router-dom";
import { Suspense, lazy } from "react";
import './App.css'

const TaskListPage = lazy(() => import("./pages/TaskListPage"));

function App() {
  return (
    <BrowserRouter>
      <Suspense fallback={<p>Loading...</p>}>
        <Routes>
          <Route path="/" element={<TaskListPage agencyId="32ea87c0-23b8-4373-9cf9-b5164bd6a675"/>} />
        </Routes>
      </Suspense>
    </BrowserRouter>
  )
}

export default App
