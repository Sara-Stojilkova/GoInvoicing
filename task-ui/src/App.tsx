import { BrowserRouter, Routes, Route } from "react-router-dom";
import { Suspense, lazy } from "react";
import './App.css'

const AGENCY_ID = "32ea87c0-23b8-4373-9cf9-b5164bd6a675";

const TaskListPage   = lazy(() => import("./pages/TaskListPage").then(m => ({ default: m.TaskListPage })));
const TaskDetailPage = lazy(() => import("./pages/TaskDetailPage").then(m => ({ default: m.TaskDetailPage })));

function App() {
  return (
    <BrowserRouter>
      <Suspense fallback={<p>Loading...</p>}>
        <Routes>
          <Route path="/"                element={<TaskListPage   agencyId={AGENCY_ID} />} />
          <Route path="/tasks/:taskId"   element={<TaskDetailPage agencyId={AGENCY_ID} />} />
        </Routes>
      </Suspense>
    </BrowserRouter>
  )
}

export default App
