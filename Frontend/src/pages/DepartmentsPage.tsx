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

function orderDepartmentsByHierarchy(departments: Department[]): Department[] {
  const byParent = new Map<number | null, Department[]>()
  for (const dept of departments) {
    const key = dept.parent_id ?? null
    const bucket = byParent.get(key)
    if (bucket) {
      bucket.push(dept)
    } else {
      byParent.set(key, [dept])
    }
  }
  for (const list of byParent.values()) {
    list.sort((a, b) => a.name.localeCompare(b.name, 'ru-RU'))
  }

  const ordered: Department[] = []
  const visited = new Set<number>()

  const walk = (parentID: number | null) => {
    const children = byParent.get(parentID) ?? []
    for (const child of children) {
      if (visited.has(child.id)) continue
      visited.add(child.id)
      ordered.push(child)
      walk(child.id)
    }
  }

  walk(null)

  // Safety: include any disconnected nodes.
  for (const dept of departments) {
    if (!visited.has(dept.id)) ordered.push(dept)
  }
  return ordered
}

function buildDepartmentOptionsHierarchical(departments: Department[]): Array<{ value: number; label: string }> {
  const byParent = new Map<number | null, Department[]>()
  for (const dept of departments) {
    const key = dept.parent_id ?? null
    const bucket = byParent.get(key)
    if (bucket) {
      bucket.push(dept)
    } else {
      byParent.set(key, [dept])
    }
  }
  for (const list of byParent.values()) {
    list.sort((a, b) => a.name.localeCompare(b.name, 'ru-RU'))
  }

  const options: Array<{ value: number; label: string }> = []
  const visited = new Set<number>()
  const walk = (parentID: number | null, depth: number) => {
    const children = byParent.get(parentID) ?? []
    for (const child of children) {
      if (visited.has(child.id)) continue
      visited.add(child.id)
      const prefix = depth > 0 ? `${'  '.repeat(depth)}└ ` : ''
      options.push({ value: child.id, label: `${prefix}${child.name}` })
      walk(child.id, depth + 1)
    }
  }

  walk(null, 0)

  for (const dept of departments) {
    if (!visited.has(dept.id)) {
      options.push({ value: dept.id, label: dept.name })
    }
  }
  return options
}

function buildDepartmentDepthMap(departments: Department[]): Map<number, number> {
  const byParent = new Map<number | null, Department[]>()
  for (const dept of departments) {
    const key = dept.parent_id ?? null
    const bucket = byParent.get(key)
    if (bucket) {
      bucket.push(dept)
    } else {
      byParent.set(key, [dept])
    }
  }

  const depthMap = new Map<number, number>()
  const visit = (parentID: number | null, depth: number) => {
    const children = byParent.get(parentID) ?? []
    children.sort((a, b) => a.name.localeCompare(b.name, 'ru-RU'))
    for (const child of children) {
      if (depthMap.has(child.id)) continue
      depthMap.set(child.id, depth)
      visit(child.id, depth + 1)
    }
  }

  visit(null, 0)

  for (const dept of departments) {
    if (!depthMap.has(dept.id)) {
      depthMap.set(dept.id, 0)
    }
  }

  return depthMap
}

export default function DepartmentsPage() {
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)
  const [parentModalOpen, setParentModalOpen] = useState(false)
  const [selectedDepartment, setSelectedDepartment] = useState<Department | null>(null)
  const [newParentId, setNewParentId] = useState<number | null>(null)
  const [form] = Form.useForm()

  const { data: departments, isLoading } = useQuery({
    queryKey: ['departments'],
    queryFn: iamApi.listDepartments,
  })

  const orderedDepartments = orderDepartmentsByHierarchy(departments ?? [])
  const departmentOptions = buildDepartmentOptionsHierarchical(departments ?? [])
  const departmentDepthMap = buildDepartmentDepthMap(departments ?? [])

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

  const setParentMutation = useMutation({
    mutationFn: (payload: { id: number; parent_id: number | null }) =>
      iamApi.setDeptParent(payload.id, payload.parent_id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['departments'] })
      message.success('Иерархия отдела обновлена')
      setParentModalOpen(false)
      setSelectedDepartment(null)
      setNewParentId(null)
    },
    onError: (e) => {
      const err = (e as { response?: { data?: { error?: string } } })?.response?.data?.error
      message.error(err ?? 'Ошибка при обновлении иерархии')
    },
  })

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    {
      title: 'Название',
      dataIndex: 'name',
      render: (name: string, record: Department) => {
        const depth = departmentDepthMap.get(record.id) ?? 0
        const prefix = depth > 0 ? `${'  '.repeat(depth)}└ ` : ''
        return `${prefix}${name}`
      },
    },
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
        <Space>
          <Button
            size="small"
            onClick={() => {
              setSelectedDepartment(record)
              setNewParentId(record.parent_id)
              setParentModalOpen(true)
            }}
          >
            Изменить иерархию
          </Button>
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
        </Space>
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
        dataSource={orderedDepartments}
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
              options={departmentOptions}
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

      <Modal
        title={selectedDepartment ? `Изменить иерархию: ${selectedDepartment.name}` : 'Изменить иерархию'}
        open={parentModalOpen}
        onCancel={() => {
          setParentModalOpen(false)
          setSelectedDepartment(null)
          setNewParentId(null)
        }}
        onOk={() => {
          if (!selectedDepartment) return
          setParentMutation.mutate({ id: selectedDepartment.id, parent_id: newParentId })
        }}
        okText="Сохранить"
        confirmLoading={setParentMutation.isPending}
      >
        <Select
          style={{ width: '100%' }}
          value={newParentId ?? undefined}
          onChange={(v) => setNewParentId(v ?? null)}
          allowClear
          placeholder="Без родителя (корневой отдел)"
          options={departmentOptions
            .filter((d) => d.value !== selectedDepartment?.id)
            .map((d) => ({ value: d.value, label: d.label }))}
        />
      </Modal>
    </div>
  )
}
