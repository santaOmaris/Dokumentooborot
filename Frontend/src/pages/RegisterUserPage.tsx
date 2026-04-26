import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Typography,
  Form,
  Input,
  Button,
  Select,
  Switch,
  message,
} from 'antd'
import { iamApi } from '../api/iam'
import type { Department } from '../types'

const { Title } = Typography

export default function RegisterUserPage() {
  const queryClient = useQueryClient()
  const [form] = Form.useForm()
  const [isHead, setIsHead] = useState(false)

  const { data: departments } = useQuery({
    queryKey: ['departments'],
    queryFn: iamApi.listDepartments,
  })

  const registerMutation = useMutation({
    mutationFn: iamApi.register,
    onSuccess: (data) => {
      message.success(`Пользователь создан, ID: ${data.user_id}`)
      form.resetFields()
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error
      message.error(msg ?? 'Ошибка при создании пользователя')
    },
  })

  return (
    <div style={{ maxWidth: 520 }}>
      <Title level={3}>Регистрация пользователя</Title>
      <Form
        form={form}
        layout="vertical"
        initialValues={{ system_role: 'USER', is_head: false }}
        onFinish={(values) =>
          registerMutation.mutate({ ...values, is_head: isHead })
        }
      >
        <Form.Item label="Логин" name="login" rules={[{ required: true }]}>
          <Input />
        </Form.Item>
        <Form.Item
          label="Email"
          name="email"
          rules={[{ required: true }, { type: 'email' }]}
        >
          <Input />
        </Form.Item>
        <Form.Item
          label="ФИО (минимум 3 слова)"
          name="full_name"
          rules={[{ required: true }]}
        >
          <Input placeholder="Иванов Иван Иванович" />
        </Form.Item>
        <Form.Item
          label="Пароль (мин. 8 символов)"
          name="password"
          rules={[{ required: true }, { min: 8 }]}
        >
          <Input.Password />
        </Form.Item>
        <Form.Item label="Системная роль" name="system_role">
          <Select
            options={[
              { value: 'USER', label: 'USER' },
              { value: 'ADMIN', label: 'ADMIN' },
            ]}
          />
        </Form.Item>
        <Form.Item label="Отдел" name="department_id">
          <Select
            options={departments?.map((d: Department) => ({ value: d.id, label: d.name }))}
            allowClear
            placeholder="Без отдела"
          />
        </Form.Item>
        <Form.Item label="Начальник отдела">
          <Switch checked={isHead} onChange={setIsHead} />
        </Form.Item>
        <Form.Item>
          <Button
            type="primary"
            htmlType="submit"
            loading={registerMutation.isPending}
            block
          >
            Создать пользователя
          </Button>
        </Form.Item>
      </Form>
    </div>
  )
}
