import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Typography,
  Tabs,
  Tag,
  Button,
  Space,
  Table,
  Input,
  Timeline,
  Descriptions,
  Modal,
  Select,
  message,
  Divider,
  Alert,
} from 'antd'
import {
  SendOutlined,
  CheckOutlined,
  CloseOutlined,
  UserSwitchOutlined,
  SwapOutlined,
  DownloadOutlined,
} from '@ant-design/icons'
import { catalogApi } from '../api/catalog'
import { collaborationApi } from '../api/collaboration'
import { orchestratorApi } from '../api/orchestrator'
import { iamApi } from '../api/iam'
import { useMe } from '../hooks/useMe'

const { Title, Text } = Typography
const { TextArea } = Input

const STATUS_COLORS: Record<string, string> = {
  DRAFT: 'default',
  PENDING_BOSS: 'processing',
  PENDING_VISA: 'warning',
  APPROVED: 'success',
  REJECTED: 'error',
  HIDDEN: 'default',
}

const STATUS_LABELS: Record<string, string> = {
  DRAFT: 'Черновик',
  PENDING_BOSS: 'На рассмотрении начальника',
  PENDING_VISA: 'На визировании',
  APPROVED: 'Согласован',
  REJECTED: 'Отклонён',
  HIDDEN: 'Скрыт',
}

export default function DocumentDetailPage() {
  const { id } = useParams<{ id: string }>()
  const docId = Number(id)
  const { data: me } = useMe()
  const queryClient = useQueryClient()
  const [newMessage, setNewMessage] = useState('')
  const [visaOpen, setVisaOpen] = useState(false)
  const [visaNote, setVisaNote] = useState('')
  const [rejectNote, setRejectNote] = useState('')
  const [rejectOpen, setRejectOpen] = useState(false)
  const [delegateOpen, setDelegateOpen] = useState(false)
  const [delegateUserId, setDelegateUserId] = useState<number | null>(null)
  const [approvalOpen, setApprovalOpen] = useState(false)
  const [approvalDeptId, setApprovalDeptId] = useState<number | null>(null)
  const [approvalQuestion, setApprovalQuestion] = useState('')

  const { data: doc } = useQuery({
    queryKey: ['doc', docId],
    queryFn: () => catalogApi.getDocument(docId),
  })

  const { data: docStatus } = useQuery({
    queryKey: ['doc-status', docId],
    queryFn: () => orchestratorApi.getStatus(docId),
  })

  const { data: history } = useQuery({
    queryKey: ['doc-history', docId],
    queryFn: () => catalogApi.getDocumentHistory(docId),
  })

  const { data: workflowHistory } = useQuery({
    queryKey: ['workflow-history', docId],
    queryFn: () => orchestratorApi.getHistory(docId),
  })

  const { data: messages, refetch: refetchMessages } = useQuery({
    queryKey: ['messages', docId],
    queryFn: () => collaborationApi.listMessages(docId),
  })

  const { data: audit } = useQuery({
    queryKey: ['doc-audit', docId],
    queryFn: () => collaborationApi.getDocAudit(docId),
  })

  const { data: users } = useQuery({
    queryKey: ['users-for-actions', me?.system_role, me?.department_id],
    enabled: !!me,
    queryFn: () => {
      const role = String(me?.system_role ?? '').toUpperCase()
      if (role === 'ADMIN') {
        return iamApi.listUsers()
      }
      const dept = Number(me?.department_id)
      return iamApi.listDeptUsers(dept)
    },
  })

  const { data: departments } = useQuery({
    queryKey: ['departments'],
    queryFn: iamApi.listDepartments,
  })

  function invalidateStatus() {
    queryClient.invalidateQueries({ queryKey: ['doc-status', docId] })
    queryClient.invalidateQueries({ queryKey: ['workflow-history', docId] })
  }

  function errMsg(e: unknown) {
    return (e as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Ошибка'
  }

  const sendMsgMutation = useMutation({
    mutationFn: () => collaborationApi.sendMessage(docId, newMessage),
    onSuccess: () => { setNewMessage(''); refetchMessages() },
  })

  const visaMutation = useMutation({
    mutationFn: () => orchestratorApi.sendForVisa(docId, { note: visaNote.trim() }),
    onSuccess: () => {
      message.success('Отправлен на визирование')
      setVisaOpen(false)
      setVisaNote('')
      invalidateStatus()
    },
    onError: (e) => message.error(errMsg(e)),
  })

  const approveMutation = useMutation({
    mutationFn: () => orchestratorApi.approve(docId),
    onSuccess: () => { message.success('Документ согласован'); invalidateStatus() },
    onError: (e) => message.error(errMsg(e)),
  })

  const rejectMutation = useMutation({
    mutationFn: () => orchestratorApi.reject(docId, rejectNote),
    onSuccess: () => {
      message.success('Документ отклонён')
      setRejectOpen(false); setRejectNote('')
      invalidateStatus()
    },
    onError: (e) => message.error(errMsg(e)),
  })

  const delegateMutation = useMutation({
    mutationFn: () => orchestratorApi.delegate(docId, delegateUserId!),
    onSuccess: () => {
      message.success('Документ делегирован')
      setDelegateOpen(false); setDelegateUserId(null)
      invalidateStatus()
    },
    onError: (e) => message.error(errMsg(e)),
  })

  const requestApprovalMutation = useMutation({
    mutationFn: () => orchestratorApi.requestApproval(docId, {
      target_department_id: approvalDeptId ?? undefined,
      question: approvalQuestion.trim(),
    }),
    onSuccess: () => {
      message.success('Запрос на согласование отправлен')
      setApprovalOpen(false)
      setApprovalDeptId(null)
      setApprovalQuestion('')
      invalidateStatus()
    },
    onError: (e) => message.error(errMsg(e)),
  })

  const status = String(docStatus?.status ?? doc?.status ?? '').trim().toUpperCase()
  const role = String(me?.system_role ?? '').toUpperCase()
  const isHead = Boolean(me?.is_head)
  const isAdmin = role === 'ADMIN'
  const canDelegate = isHead || isAdmin
  const canRequestApproval = (isHead || isAdmin) && ['PENDING_VISA', 'PENDING_BOSS'].includes(status)

  const historyColumns = [
    { title: 'Событие', dataIndex: 'change_type' },
    { title: 'Описание', dataIndex: 'description' },
    { title: 'Когда', dataIndex: 'created_at', render: (v: string) => new Date(v).toLocaleString('ru-RU') },
  ]
  const auditColumns = [
    { title: 'Действие', dataIndex: 'action' },
    { title: 'Кто', dataIndex: 'actor_login' },
    { title: 'Детали', dataIndex: 'details' },
    { title: 'Когда', dataIndex: 'created_at', render: (v: string) => new Date(v).toLocaleString('ru-RU') },
  ]

  return (
    <div>
      <Title level={3}>{doc?.title ?? '...'}</Title>

      <Descriptions bordered size="small" style={{ marginBottom: 16 }}>
        <Descriptions.Item label="Статус">
          <Tag color={STATUS_COLORS[status] ?? 'default'}>{STATUS_LABELS[status] ?? status}</Tag>
        </Descriptions.Item>
        <Descriptions.Item label="Создан">
          {doc?.created_at ? new Date(doc.created_at).toLocaleString('ru-RU') : '—'}
        </Descriptions.Item>
        {docStatus?.assignee_id && (
          <Descriptions.Item label="Исполнитель">
            {users?.find(u => u.id === docStatus.assignee_id)?.full_name ?? `ID ${docStatus.assignee_id}`}
          </Descriptions.Item>
        )}
      </Descriptions>

      {status === 'REJECTED' && (
        <Alert message="Документ отклонён и требует доработки" type="error" showIcon style={{ marginBottom: 16 }} />
      )}

      <Space wrap style={{ marginBottom: 16 }}>
        {status === 'DRAFT' && (
          <Button type="primary" icon={<SendOutlined />}
            onClick={() => setVisaOpen(true)}>
            На визирование
          </Button>
        )}
        <Button
          icon={<UserSwitchOutlined />}
          onClick={() => setDelegateOpen(true)}
          disabled={!canDelegate || ['APPROVED', 'HIDDEN'].includes(status)}
          title={!canDelegate ? 'Доступно только начальнику отдела или администратору' : undefined}
        >
          Делегировать
        </Button>
        <Button
          icon={<SwapOutlined />}
          onClick={() => setApprovalOpen(true)}
          loading={requestApprovalMutation.isPending}
          disabled={!canRequestApproval}
          title={!canRequestApproval ? 'Доступно в статусе PENDING_VISA/PENDING_BOSS для начальника/админа' : undefined}
        >
          Запросить согласование
        </Button>
        {isHead && ['PENDING_VISA'].includes(status) && (
          <>
            <Button type="primary" icon={<CheckOutlined />}
              onClick={() => approveMutation.mutate()} loading={approveMutation.isPending}>
              Согласовать
            </Button>
            <Button danger icon={<CloseOutlined />} onClick={() => setRejectOpen(true)}>
              Отклонить
            </Button>
          </>
        )}
        <Button icon={<DownloadOutlined />}
          onClick={() => {
            catalogApi.downloadDocument(docId)
              .then(blob => {
                const url = URL.createObjectURL(blob)
                const a = document.createElement('a')
                a.href = url; a.download = doc?.title ?? 'document'
                a.click(); URL.revokeObjectURL(url)
              })
              .catch(() => message.error('Ошибка при скачивании'))
          }}>
          Скачать
        </Button>
      </Space>

      <Divider />

      <Tabs items={[
        {
          key: 'chat', label: 'Обсуждение',
          children: (
            <div>
              <div style={{ maxHeight: 320, overflowY: 'auto', marginBottom: 16 }}>
                {!messages?.length && <Text type="secondary">Сообщений пока нет</Text>}
                {messages?.map(m => (
                  <div key={m.id} style={{ marginBottom: 8, padding: '6px 10px', background: '#f5f5f5', borderRadius: 6 }}>
                    <Text strong>{m.sender_login}: </Text><Text>{m.content}</Text>
                    <Text type="secondary" style={{ marginLeft: 8, fontSize: 11 }}>
                      {new Date(m.created_at).toLocaleString('ru-RU')}
                    </Text>
                  </div>
                ))}
              </div>
              <Space.Compact style={{ width: '100%' }}>
                <Input value={newMessage} onChange={e => setNewMessage(e.target.value)}
                  placeholder="Написать сообщение..." onPressEnter={() => newMessage && sendMsgMutation.mutate()} />
                <Button type="primary" onClick={() => sendMsgMutation.mutate()}
                  loading={sendMsgMutation.isPending} disabled={!newMessage}>Отправить</Button>
              </Space.Compact>
            </div>
          ),
        },
        {
          key: 'history', label: 'История изменений',
          children: <Table dataSource={history} columns={historyColumns} rowKey="id" pagination={false} size="small" />,
        },
        {
          key: 'workflow', label: 'Маршрут согласования',
          children: (
            <Timeline items={workflowHistory?.map(e => ({
              color: e.to_state === 'APPROVED' ? 'green' : e.to_state === 'REJECTED' ? 'red' : 'blue',
              children: (
                <div>
                  <Text strong>{e.from_state} → {e.to_state}</Text><br />
                  <Text type="secondary">{e.triggered_by} · {new Date(e.created_at).toLocaleString('ru-RU')}</Text>
                  {e.note && <><br /><Text italic>{e.note}</Text></>}
                </div>
              ),
            }))} />
          ),
        },
        {
          key: 'audit', label: 'Журнал аудита действий',
          children: <Table dataSource={audit} columns={auditColumns} rowKey="id" pagination={false} size="small" />,
        },
      ]} />

      <Modal title="Комментарий к визированию" open={visaOpen}
        onCancel={() => { setVisaOpen(false); setVisaNote('') }}
        onOk={() => visaMutation.mutate()} confirmLoading={visaMutation.isPending}
        okText="Отправить"
        okButtonProps={{ disabled: !visaNote.trim() }}>
        <TextArea
          rows={4}
          value={visaNote}
          onChange={e => setVisaNote(e.target.value)}
          placeholder="Укажите комментарий для визирующего"
        />
      </Modal>

      <Modal title="Причина отклонения" open={rejectOpen}
        onCancel={() => { setRejectOpen(false); setRejectNote('') }}
        onOk={() => rejectMutation.mutate()} confirmLoading={rejectMutation.isPending}
        okText="Отклонить" okButtonProps={{ danger: true }}>
        <TextArea rows={4} value={rejectNote} onChange={e => setRejectNote(e.target.value)}
          placeholder="Укажите что нужно исправить..." />
      </Modal>

      <Modal title="Делегировать документ" open={delegateOpen}
        onCancel={() => { setDelegateOpen(false); setDelegateUserId(null) }}
        onOk={() => delegateUserId && delegateMutation.mutate()}
        confirmLoading={delegateMutation.isPending}
        okText="Делегировать" okButtonProps={{ disabled: !delegateUserId }}>
        <Select showSearch placeholder="Выберите сотрудника" style={{ width: '100%' }}
          value={delegateUserId} onChange={setDelegateUserId}
          filterOption={(input, option) =>
            String(option?.label ?? '').toLowerCase().includes(input.toLowerCase())
          }
          options={users?.filter(u => u.id !== me?.user_id)
            .map(u => ({ value: u.id, label: `${u.full_name} (${u.login})` }))} />
      </Modal>

      <Modal
        title="Запросить согласование в отделе"
        open={approvalOpen}
        onCancel={() => {
          setApprovalOpen(false)
          setApprovalDeptId(null)
          setApprovalQuestion('')
        }}
        onOk={() => requestApprovalMutation.mutate()}
        confirmLoading={requestApprovalMutation.isPending}
        okText="Отправить"
        okButtonProps={{ disabled: !approvalDeptId || !approvalQuestion.trim() }}
      >
        <Select
          showSearch
          placeholder="Выберите отдел"
          style={{ width: '100%' }}
          value={approvalDeptId}
          onChange={setApprovalDeptId}
          options={departments?.map((d) => ({ value: d.id, label: d.name }))}
        />
        <TextArea
          rows={4}
          value={approvalQuestion}
          onChange={e => setApprovalQuestion(e.target.value)}
          style={{ marginTop: 12 }}
          placeholder="Добавьте комментарий для отдела"
        />
      </Modal>
    </div>
  )
}
