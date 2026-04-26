import { useQuery } from '@tanstack/react-query'
import { Typography, Table, Select, Space } from 'antd'
import { useState } from 'react'
import { iamApi } from '../api/iam'
import { collaborationApi } from '../api/collaboration'
import { useMe } from '../hooks/useMe'
import type { Department } from '../types'

const { Title } = Typography

export default function AuditPage() {
  const { data: me } = useMe()
  const isAdmin = me?.system_role === 'ADMIN'
  const myDept = me?.department_id ? Number(me.department_id) : null
  const [selectedDept, setSelectedDept] = useState<number | null>(
    me?.department_id ? Number(me.department_id) : null
  )
  const effectiveDept = isAdmin ? selectedDept : myDept

  const { data: departments } = useQuery({
    queryKey: ['departments'],
    queryFn: iamApi.listDepartments,
  })

  const { data: logs, isLoading } = useQuery({
    queryKey: ['dept-audit', effectiveDept],
    queryFn: () => collaborationApi.getDeptAudit(effectiveDept!),
    enabled: !!effectiveDept,
  })

  const columns = [
    {
      title: 'Действие',
      dataIndex: 'action',
      width: 200,
    },
    {
      title: 'Кто',
      dataIndex: 'actor_login',
      width: 140,
    },
    {
      title: 'Документ',
      dataIndex: 'document_id',
      width: 100,
      render: (v: number | null) => v ?? '—',
    },
    {
      title: 'Детали',
      dataIndex: 'details',
    },
    {
      title: 'Когда',
      dataIndex: 'created_at',
      width: 170,
      render: (v: string) => new Date(v).toLocaleString('ru-RU'),
    },
  ]

  return (
    <div>
      <Title level={3}>Журнал действий отдела</Title>

      {isAdmin ? (
        <Space style={{ marginBottom: 16 }}>
          <Select
            placeholder="Выберите отдел"
            style={{ width: 280 }}
            value={selectedDept}
            onChange={setSelectedDept}
            options={departments?.map((d: Department) => ({ value: d.id, label: d.name }))}
            allowClear
          />
        </Space>
      ) : (
        <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 16 }}>
          Журнал вашего отдела
        </Typography.Text>
      )}

      <Table
        dataSource={logs}
        columns={columns}
        rowKey="id"
        loading={isLoading}
        pagination={{ pageSize: 30 }}
        size="small"
      />
    </div>
  )
}
