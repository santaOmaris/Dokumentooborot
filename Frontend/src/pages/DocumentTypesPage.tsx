import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Typography,
  Table,
  Button,
  Space,
  Input,
  Popconfirm,
  message,
} from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useState } from 'react'
import { catalogApi } from '../api/catalog'
import type { DocumentType } from '../types'

const { Title } = Typography

export default function DocumentTypesPage() {
  const queryClient = useQueryClient()
  const [newName, setNewName] = useState('')

  const { data: types, isLoading } = useQuery({
    queryKey: ['types'],
    queryFn: catalogApi.listTypes,
  })

  const createMutation = useMutation({
    mutationFn: () => catalogApi.createType(newName.trim()),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['types'] })
      message.success('Тип создан')
      setNewName('')
    },
    onError: () => message.error('Ошибка при создании'),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => catalogApi.deleteType(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['types'] })
      message.success('Тип удалён')
    },
    onError: () => message.error('Ошибка при удалении'),
  })

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: 'Название', dataIndex: 'name' },
    {
      title: 'Действия',
      render: (_: unknown, record: DocumentType) => (
        <Popconfirm
          title="Удалить тип?"
          onConfirm={() => deleteMutation.mutate(record.id)}
          okText="Удалить" okButtonProps={{ danger: true }}
        >
          <Button size="small" danger loading={deleteMutation.isPending}>Удалить</Button>
        </Popconfirm>
      ),
    },
  ]

  return (
    <div>
      <Title level={3}>Типы документов</Title>

      <Space style={{ marginBottom: 16 }}>
        <Input
          placeholder="Название нового типа"
          value={newName}
          onChange={e => setNewName(e.target.value)}
          onPressEnter={() => newName.trim() && createMutation.mutate()}
          style={{ width: 300 }}
        />
        <Button
          type="primary"
          icon={<PlusOutlined />}
          disabled={!newName.trim()}
          loading={createMutation.isPending}
          onClick={() => createMutation.mutate()}
        >
          Добавить тип
        </Button>
      </Space>

      <Table
        dataSource={types}
        columns={columns}
        rowKey="id"
        loading={isLoading}
        pagination={false}
      />
    </div>
  )
}
