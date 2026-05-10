import { useState } from 'react'
import { Layout, Menu, Avatar, Button, Typography, Dropdown } from 'antd'
import {
  FileTextOutlined,
  TeamOutlined,
  BankOutlined,
  DashboardOutlined,
  UserOutlined,
  LogoutOutlined,
  UserAddOutlined,
  AuditOutlined,
  FileProtectOutlined,
} from '@ant-design/icons'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { useMe } from '../hooks/useMe'
import { iamApi } from '../api/iam'

const { Header, Sider, Content } = Layout
const { Text } = Typography

export default function AppLayout() {
  const { data: me } = useMe()
  const navigate = useNavigate()
  const location = useLocation()
  const queryClient = useQueryClient()
  const [collapsed, setCollapsed] = useState(false)

  async function handleLogout() {
    await iamApi.logout()
    queryClient.clear()
    navigate('/login')
  }

  const menuItems = [
    {
      key: '/',
      icon: <DashboardOutlined />,
      label: 'Главная',
    },
    {
      key: '/documents',
      icon: <FileTextOutlined />,
      label: 'Документы',
    },
    ...(me?.is_head || me?.system_role === 'ADMIN'
      ? [
          {
            key: '/users',
            icon: <TeamOutlined />,
            label: 'Сотрудники',
          },
        ]
      : []),
    ...(me?.system_role === 'ADMIN'
      ? [
          {
            key: '/departments',
            icon: <BankOutlined />,
            label: 'Отделы',
          },
          {
            key: '/doc-types',
            icon: <FileProtectOutlined />,
            label: 'Типы документов',
          },
          {
            key: '/admin/register',
            icon: <UserAddOutlined />,
            label: 'Регистрация',
          },
        ]
      : []),
    {
      key: '/audit',
      icon: <AuditOutlined />,
      label: 'Журнал аудита',
    },
  ]

  const userMenu = {
    items: [
      {
        key: 'logout',
        icon: <LogoutOutlined />,
        label: 'Выйти',
        onClick: handleLogout,
      },
    ],
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider collapsible collapsed={collapsed} onCollapse={setCollapsed}>
        <div style={{ padding: '16px', textAlign: 'center' }}>
          <Text strong style={{ color: '#fff', fontSize: collapsed ? 12 : 18 }}>
            {collapsed ? 'DF' : 'DocFlow'}
          </Text>
        </div>
        <Menu
          theme="dark"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Header
          style={{
            background: '#fff',
            padding: '0 24px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'flex-end',
            borderBottom: '1px solid #f0f0f0',
          }}
        >
          <Dropdown menu={userMenu} placement="bottomRight">
            <Button type="text" style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <Avatar icon={<UserOutlined />} size="small" />
              <span>{me?.login}</span>
            </Button>
          </Dropdown>
        </Header>
        <Content style={{ margin: 24, background: '#fff', padding: 24, borderRadius: 8 }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
