import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Typography,
  Table,
  Button,
  Space,
  Tag,
  Popconfirm,
  message,
  Select,
} from 'antd'
import { iamApi } from '../api/iam'
import { useMe } from '../hooks/useMe'
import type { User, Department } from '../types'

const { Title } = Typography

export default function UsersPage() {
  const { data: me } = useMe()
  const queryClient = useQueryClient()
  const isAdmin = me?.system_role === 'ADMIN'
  const isHead = Boolean(me?.is_head)
  const myDeptId = me?.department_id ? Number(me.department_id) : null

  const { data: users, isLoading } = useQuery({
    queryKey: ['users', isAdmin ? 'all' : `dept-${myDeptId ?? 'none'}`],
    queryFn: () => {
      if (isAdmin) return iamApi.listUsers()
      if (myDeptId) return iamApi.listDeptUsers(myDeptId)
      return Promise.resolve([])
    },
    enabled: Boolean(me),
  })

  const { data: departments } = useQuery({
    queryKey: ['departments'],
    queryFn: iamApi.listDepartments,
  })

  const fireMutation = useMutation({
    mutationFn: (id: number) => iamApi.fireUser(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      message.success('Пользователь уволен')
    },
    onError: () => message.error('Ошибка при увольнении'),
  })

  const moveMutation = useMutation({
    mutationFn: ({ user_id, department_id }: { user_id: number; department_id: number }) =>
      iamApi.moveUser(user_id, department_id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      message.success('Сотрудник переведён')
    },
    onError: () => message.error('Ошибка при переводе'),
  })

  const promoteMutation = useMutation({
    mutationFn: (id: number) => iamApi.promoteUser(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      message.success('Назначен начальником')
    },
    onError: () => message.error('Ошибка'),
  })

  const demoteMutation = useMutation({
    mutationFn: (id: number) => iamApi.demoteUser(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      message.success('Разжалован')
    },
    onError: () => message.error('Ошибка'),
  })

  const visibleUsers = (users ?? []).filter((u) => {
    if (isAdmin) return true
    if (!myDeptId) return false
    return u.department_id === myDeptId
  })

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: 'Логин', dataIndex: 'login' },
    { title: 'ФИО', dataIndex: 'full_name' },
    {
      title: 'Должность',
      render: (_: unknown, r: User) => (
        <Space size="small">
          {r.is_head && <Tag color="gold">Начальник</Tag>}
          <Tag>{r.system_role}</Tag>
        </Space>
      ),
    },
    {
      title: 'Отдел',
      dataIndex: 'department_id',
      render: (pid: number | null) => {
        if (!pid) return '—'
        const name = departments?.find((d: Department) => d.id === pid)?.name
        return name ? `${pid} · ${name}` : String(pid)
      },
    },
    ...(isAdmin
      ? [
          {
            title: 'Перевести в отдел',
            render: (_: unknown, record: User) => (
              <Select
                style={{ width: 180 }}
                placeholder="Выбрать отдел"
                size="small"
                onSelect={(deptId: number) =>
                  moveMutation.mutate({ user_id: record.id, department_id: deptId })
                }
                options={departments?.map((d: Department) => ({ value: d.id, label: d.name }))}
              />
            ),
          },
          {
            title: 'Действия',
            render: (_: unknown, record: User) => (
              <Space size="small">
                {!record.is_head && (
                  <Button
                    size="small"
                    onClick={() => promoteMutation.mutate(record.id)}
                    loading={promoteMutation.isPending}
                  >
                    Назначить нач.
                  </Button>
                )}
                {record.is_head && (
                  <Button
                    size="small"
                    onClick={() => demoteMutation.mutate(record.id)}
                    loading={demoteMutation.isPending}
                  >
                    Разжаловать
                  </Button>
                )}
                <Popconfirm
                  title="Уволить сотрудника?"
                  onConfirm={() => fireMutation.mutate(record.id)}
                  okText="Уволить"
                  okButtonProps={{ danger: true }}
                >
                  <Button size="small" danger>
                    Уволить
                  </Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]
      : isHead
        ? [
            {
              title: 'Действия',
              render: (_: unknown, record: User) => (
                <Space size="small">
                  <Popconfirm
                    title="Уволить сотрудника?"
                    onConfirm={() => fireMutation.mutate(record.id)}
                    okText="Уволить"
                    okButtonProps={{ danger: true }}
                  >
                    <Button size="small" danger loading={fireMutation.isPending} disabled={record.id === me?.user_id}>
                      Уволить
                    </Button>
                  </Popconfirm>
                </Space>
              ),
            },
          ]
        : []),
  ]

  return (
    <div>
      <Title level={3}>Управление сотрудниками</Title>
      <Table
        dataSource={visibleUsers}
        columns={columns}
        rowKey="id"
        loading={isLoading}
        pagination={{ pageSize: 20 }}
      />
    </div>
  )
}
