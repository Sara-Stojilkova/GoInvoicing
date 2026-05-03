import { BrowserRouter, Routes, Route } from "react-router-dom";
import { Suspense, lazy } from "react";
import "./App.css";
import { Header } from "./component/Header";
import { AuthProvider, useAuth } from "./context/AuthContext";
import { ProtectedRoute } from "./component/ProtectedRoute";
import { LoginPage } from "./pages/LoginPage";
import { RegisterPage } from "./pages/RegisterPage";

const TaskListPage   = lazy(() => import("./pages/TaskListPage").then((m) => ({ default: m.TaskListPage })));
const TaskDetailPage = lazy(() => import("./pages/TaskDetailPage").then((m) => ({ default: m.TaskDetailPage })));

function AppRoutes() {
  const { agencyId } = useAuth();
  return (
    <Suspense fallback={<p>Loading…</p>}>
      <Routes>
        <Route path="/login"    element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route element={<ProtectedRoute />}>
          <Route path="/"              element={<TaskListPage   agencyId={agencyId ?? ""} />} />
          <Route path="/tasks/:taskId" element={<TaskDetailPage agencyId={agencyId ?? ""} />} />
        </Route>
      </Routes>
    </Suspense>
  );
}

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Header />
        <AppRoutes />
      </AuthProvider>
    </BrowserRouter>
  );
}

export default App;
