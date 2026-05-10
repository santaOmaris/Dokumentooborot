import { useState } from 'react'
import { Form, Input, Button, Card, Typography, message } from 'antd'
import { useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { iamApi } from '../api/iam'

const { Title } = Typography

export default function LoginPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [loading, setLoading] = useState(false)

  async function onFinish(values: { login: string; password: string }) {
    setLoading(true)
    try {
      await iamApi.login(values.login, values.password)
      await queryClient.invalidateQueries({ queryKey: ['me'] })
      navigate('/')
    } catch {
      message.error('Неверный логин или пароль')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#f0f2f5' }}>
      <Card style={{ width: 380 }}>
        <Title level={3} style={{ textAlign: 'center', marginBottom: 24 }}>DocFlow</Title>
        <Form layout="vertical" onFinish={onFinish}>
          <Form.Item label="Логин" name="login" rules={[{ required: true }]}>
            <Input size="large" autoFocus />
          </Form.Item>
          <Form.Item label="Пароль" name="password" rules={[{ required: true }]}>
            <Input.Password size="large" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block size="large">
              Войти
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
