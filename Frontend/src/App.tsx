import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import RequireAuth from './layouts/RequireAuth'
import AppLayout from './layouts/AppLayout'
import LoginPage from './pages/LoginPage'
import DashboardPage from './pages/DashboardPage'
import DocumentsPage from './pages/DocumentsPage'
import DocumentDetailPage from './pages/DocumentDetailPage'
import DepartmentsPage from './pages/DepartmentsPage'
import UsersPage from './pages/UsersPage'
import RegisterUserPage from './pages/RegisterUserPage'
import DocumentTypesPage from './pages/DocumentTypesPage'
import AuditPage from './pages/AuditPage'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />

        <Route element={<RequireAuth />}>
          <Route element={<AppLayout />}>
            <Route path="/" element={<DashboardPage />} />
            <Route path="/documents" element={<DocumentsPage />} />
            <Route path="/documents/:id" element={<DocumentDetailPage />} />
            <Route path="/departments" element={<DepartmentsPage />} />
            <Route path="/users" element={<UsersPage />} />
            <Route path="/admin/register" element={<RegisterUserPage />} />
            <Route path="/doc-types" element={<DocumentTypesPage />} />
            <Route path="/audit" element={<AuditPage />} />
          </Route>
        </Route>

        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
