import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Typography,
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
  Space,
  message,
  Popconfirm,
} from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { iamApi } from '../api/iam'
import type { Department } from '../types'

const { Title } = Typography

export default function DepartmentsPage() {
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)
  const [form] = Form.useForm()

  const { data: departments, isLoading } = useQuery({
    queryKey: ['departments'],
    queryFn: iamApi.listDepartments,
  })

  const createMutation = useMutation({
    mutationFn: (values: { name: string; parent_id?: number }) =>
      iamApi.createDepartment(values.name, values.parent_id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['departments'] })
      message.success('Отдел создан')
      setOpen(false)
      form.resetFields()
    },
    onError: () => message.error('Ошибка при создании'),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => iamApi.deleteDepartment(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['departments'] })
      message.success('Отдел удалён')
    },
    onError: () => message.error('Ошибка при удалении'),
  })

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: 'Название', dataIndex: 'name' },
    {
      title: 'Родительский отдел',
      dataIndex: 'parent_id',
      render: (pid: number | null) => {
        if (!pid) return '—'
        return departments?.find((d: Department) => d.id === pid)?.name ?? pid
      },
    },
    {
      title: 'Действия',
      render: (_: unknown, record: Department) => (
        <Popconfirm
          title="Удалить отдел?"
          onConfirm={() => deleteMutation.mutate(record.id)}
          okText="Удалить"
          okButtonProps={{ danger: true }}
        >
          <Button size="small" danger loading={deleteMutation.isPending}>
            Удалить
          </Button>
        </Popconfirm>
      ),
    },
  ]

  return (
    <div>
      <Space style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={3} style={{ margin: 0 }}>Управление отделами</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setOpen(true)}>
          Создать отдел
        </Button>
      </Space>

      <Table
        dataSource={departments}
        columns={columns}
        rowKey="id"
        loading={isLoading}
        pagination={{ pageSize: 20 }}
      />

      <Modal
        title="Новый отдел"
        open={open}
        onCancel={() => setOpen(false)}
        footer={null}
      >
        <Form form={form} layout="vertical" onFinish={(v) => createMutation.mutate(v)}>
          <Form.Item label="Название" name="name" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item label="Родительский отдел" name="parent_id">
            <Select
              options={departments?.map((d: Department) => ({ value: d.id, label: d.name }))}
              placeholder="Без родителя (корневой)"
              allowClear
            />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={createMutation.isPending} block>
              Создать
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
