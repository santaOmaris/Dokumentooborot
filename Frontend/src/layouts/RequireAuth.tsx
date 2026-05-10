import { Navigate, Outlet } from 'react-router-dom'
import { Spin } from 'antd'
import { useMe } from '../hooks/useMe'

export default function RequireAuth() {
  const { data: me, isLoading, isError } = useMe()

  if (isLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh' }}>
        <Spin size="large" />
      </div>
    )
  }

  if (isError || !me) {
    return <Navigate to="/login" replace />
  }

  return <Outlet />
}
