import { useEffect, useMemo, useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Table,
  Button,
  Space,
  Typography,
  Select,
  Tabs,
  Tag,
  Upload,
  Modal,
  Form,
  Input,
  message,
} from 'antd'
import { UploadOutlined, SearchOutlined, FolderOpenOutlined } from '@ant-design/icons'
import { Link } from 'react-router-dom'
import { useMe } from '../hooks/useMe'
import { catalogApi } from '../api/catalog'
import { iamApi } from '../api/iam'
import { orchestratorApi } from '../api/orchestrator'
import type { Document, Folder } from '../types'

const { Title, Text } = Typography

const STATUS_COLORS: Record<string, string> = {
  DRAFT: 'default',
  PENDING_BOSS: 'processing',
  PENDING_VISA: 'warning',
  APPROVED: 'success',
  REJECTED: 'error',
  HIDDEN: 'default',
}

const SYSTEM_FOLDER_LABELS: Record<string, string> = {
  archived: 'Архив',
  collaborations: 'Коллаборации',
  head_only: 'Для начальника',
  main: 'Основная',
  templates: 'Шаблоны',
}

type FolderOption = {
  id: number
  name: string
}

function isHeadOnlyFolder(folder: Folder): boolean {
  return Boolean(folder.is_system) && folder.name === 'head_only'
}

function displayFolderName(folder: Folder): string {
  const translated = SYSTEM_FOLDER_LABELS[folder.name.toLowerCase()]
  return translated ?? folder.name
}

function formatRuDate(value?: string): string {
  if (!value) return '—'
  const dt = new Date(value)
  return Number.isNaN(dt.getTime()) ? '—' : dt.toLocaleDateString('ru-RU')
}

function apiErr(e: unknown, fallback: string): string {
  return (e as { response?: { data?: { error?: string } } })?.response?.data?.error ?? fallback
}

export default function DocumentsPage() {
  const { data: me } = useMe()
  const myDeptId = me ? Number(me.department_id) : 0
  const role = String(me?.system_role ?? '').toUpperCase()
  const isAdmin = role === 'ADMIN'
  const isHead = Boolean(me?.is_head)
  const canSeeReviewQueue = isHead || isAdmin
  const canAccessHeadOnly = Boolean(me?.is_head || role === 'ADMIN')

  const [selectedDeptId, setSelectedDeptId] = useState<number | null>(null)
  const [selectedFolderId, setSelectedFolderId] = useState<number | null>(null)
  const [folderSearchText, setFolderSearchText] = useState('')
  const [searchText, setSearchText] = useState('')
  const [uploadOpen, setUploadOpen] = useState(false)
  const [createFolderOpen, setCreateFolderOpen] = useState(false)
  const [newFolderName, setNewFolderName] = useState('')
  const [newFolderParentId, setNewFolderParentId] = useState<number | null>(null)
  const [moveOpen, setMoveOpen] = useState(false)
  const [moveDoc, setMoveDoc] = useState<Document | null>(null)
  const [moveFolderId, setMoveFolderId] = useState<number | undefined>(undefined)
  const [form] = Form.useForm()
  const queryClient = useQueryClient()

  const { data: departments } = useQuery({
    queryKey: ['departments-for-docs', isAdmin],
    queryFn: iamApi.listDepartments,
    enabled: isAdmin,
  })

  useEffect(() => {
    if (!isAdmin || selectedDeptId) return
    if (!departments?.length) return
    const preferred = departments.find((d) => d.id === myDeptId)
    setSelectedDeptId(preferred?.id ?? departments[0].id)
  }, [isAdmin, selectedDeptId, departments, myDeptId])

  const effectiveDeptId = isAdmin ? (selectedDeptId ?? 0) : myDeptId

  const { data: folders, isLoading: foldersLoading } = useQuery({
    queryKey: ['folders', effectiveDeptId],
    queryFn: () => catalogApi.listFolders(effectiveDeptId),
    enabled: !!effectiveDeptId,
  })

  const visibleFolders = useMemo(
    () => (folders ?? []).filter((f) => canAccessHeadOnly || !isHeadOnlyFolder(f)),
    [folders, canAccessHeadOnly]
  )

  const folderMap = useMemo(() => {
    const map = new Map<number, Folder>()
    visibleFolders.forEach((f) => map.set(f.id, f))
    return map
  }, [visibleFolders])

  useEffect(() => {
    if (!visibleFolders.length) {
      setSelectedFolderId(null)
      return
    }
    if (selectedFolderId && folderMap.has(selectedFolderId)) {
      return
    }
    const mainFolder = visibleFolders.find((f) => f.is_system && f.name === 'main')
    setSelectedFolderId(mainFolder?.id ?? visibleFolders[0].id)
  }, [visibleFolders, selectedFolderId, folderMap])

  const selectedFolder = useMemo(
    () => (selectedFolderId ? folderMap.get(selectedFolderId) ?? null : null),
    [selectedFolderId, folderMap]
  )

  const parentFolder = useMemo(() => {
    if (!selectedFolder?.parent_id) return null
    return folderMap.get(selectedFolder.parent_id) ?? null
  }, [selectedFolder, folderMap])

  const systemFolderOptions = useMemo<FolderOption[]>(() => {
    const query = folderSearchText.trim().toLowerCase()
    return visibleFolders
      .filter((f) => Boolean(f.is_system))
      .map((f) => ({ id: f.id, name: displayFolderName(f) }))
      .filter((f) => !query || f.name.toLowerCase().includes(query))
      .sort((a, b) => a.name.localeCompare(b.name, 'ru-RU'))
  }, [visibleFolders, folderSearchText])

  const childFolders = useMemo(() => {
    if (!selectedFolderId) return []
    const query = folderSearchText.trim().toLowerCase()
    return visibleFolders
      .filter((f) => f.parent_id === selectedFolderId && !f.is_system)
      .filter((f) => !query || f.name.toLowerCase().includes(query))
      .sort((a, b) => a.name.localeCompare(b.name, 'ru-RU'))
  }, [visibleFolders, selectedFolderId, folderSearchText])

  const availableFoldersForActions = useMemo(
    () => visibleFolders
      .map((f) => ({ value: f.id, label: displayFolderName(f) }))
      .sort((a, b) => a.label.localeCompare(b.label, 'ru-RU')),
    [visibleFolders]
  )

  const { data: types } = useQuery({
    queryKey: ['types'],
    queryFn: catalogApi.listTypes,
  })

  const { data: docs, isLoading: docsLoading } = useQuery({
    queryKey: ['docs-by-folder', selectedFolderId],
    queryFn: async () => {
      if (!selectedFolderId) return []
      return catalogApi.listDocumentsByFolder(selectedFolderId)
    },
    enabled: !!selectedFolderId,
  })

  const { data: searchResults, isFetching: searching } = useQuery({
    queryKey: ['search', effectiveDeptId, searchText],
    queryFn: () => catalogApi.searchDocuments(effectiveDeptId, searchText),
    enabled: !!searchText && !!effectiveDeptId,
  })

  const { data: localDeptDocs, isFetching: localSearching } = useQuery({
    queryKey: ['local-dept-docs', effectiveDeptId],
    queryFn: async () => {
      const deptFolders = await catalogApi.listFolders(effectiveDeptId)
      const all = await Promise.all(deptFolders.map((f) => catalogApi.listDocumentsByFolder(f.id).catch(() => [])))
      const unique = new Map<number, Document>()
      all.flat().forEach((d) => unique.set(d.id, d))
      return Array.from(unique.values())
    },
    enabled: !!effectiveDeptId && (!!searchText || canSeeReviewQueue),
  })

  const localDocIdsKey = useMemo(
    () => (localDeptDocs ?? []).map((d) => d.id).sort((a, b) => a - b).join(','),
    [localDeptDocs]
  )

  const { data: myReviewQueueDocs, isFetching: reviewQueueLoading } = useQuery({
    queryKey: ['review-queue', effectiveDeptId, me?.user_id, localDocIdsKey],
    enabled: canSeeReviewQueue && !!effectiveDeptId && !!me?.user_id && !!localDeptDocs?.length,
    queryFn: async () => {
      const docs = localDeptDocs ?? []
      const myUserId = Number(me?.user_id ?? 0)
      if (!myUserId || docs.length === 0) return [] as Document[]

      const docsWithStatus = await Promise.all(
        docs.map(async (d) => {
          try {
            const st = await orchestratorApi.getStatus(d.id)
            return {
              ...d,
              status: String(st.status ?? '').toUpperCase(),
              assignee_id: st.assignee_id ?? d.assignee_id,
            }
          } catch {
            return {
              ...d,
              status: String(d.status ?? '').toUpperCase(),
            }
          }
        })
      )

      return docsWithStatus
        .filter((d) => {
          const assigneeId = Number(d.assignee_id ?? 0)
          return assigneeId === myUserId && (d.status === 'PENDING_VISA' || d.status === 'PENDING_BOSS')
        })
        .sort((a, b) => {
          const left = new Date(a.updated_at ?? a.created_at ?? '').getTime()
          const right = new Date(b.updated_at ?? b.created_at ?? '').getTime()
          return right - left
        })
    },
  })

  const normalizedReviewQueue = useMemo(() => {
    if (myReviewQueueDocs?.length) return myReviewQueueDocs
    const myUserId = Number(me?.user_id ?? 0)
    if (!myUserId) return [] as Document[]

    return (localDeptDocs ?? [])
      .filter((d) => {
        const assigneeId = Number(d.assignee_id ?? 0)
        const status = String(d.status ?? '').toUpperCase()
        return assigneeId === myUserId && (status === 'PENDING_VISA' || status === 'PENDING_BOSS')
      })
      .sort((a, b) => {
        const left = new Date(a.updated_at ?? a.created_at ?? '').getTime()
        const right = new Date(b.updated_at ?? b.created_at ?? '').getTime()
        return right - left
      })
  }, [myReviewQueueDocs, localDeptDocs, me?.user_id])

  const filteredReviewQueueDocs = useMemo(() => {
    const query = searchText.trim().toLowerCase()
    if (!query) return normalizedReviewQueue
    return normalizedReviewQueue.filter((d) =>
      d.title?.toLowerCase().includes(query) || d.original_name?.toLowerCase().includes(query)
    )
  }, [normalizedReviewQueue, searchText])

  const uploadMutation = useMutation({
    mutationFn: (fd: FormData) => catalogApi.uploadDocument(fd),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['docs-by-folder'] })
      queryClient.invalidateQueries({ queryKey: ['search', effectiveDeptId] })
      queryClient.invalidateQueries({ queryKey: ['local-dept-docs', effectiveDeptId] })
      queryClient.invalidateQueries({ queryKey: ['folders', effectiveDeptId] })
      message.success('Документ загружен')
      setUploadOpen(false)
      form.resetFields()
    },
    onError: () => message.error('Ошибка при загрузке'),
  })

  const hideMutation = useMutation({
    mutationFn: (id: number) => catalogApi.hideDocument(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['docs-by-folder'] })
      queryClient.invalidateQueries({ queryKey: ['search', effectiveDeptId] })
      queryClient.invalidateQueries({ queryKey: ['local-dept-docs', effectiveDeptId] })
      message.success('Документ скрыт')
    },
    onError: () => message.error('Недостаточно прав для скрытия документа'),
  })

  const unhideMutation = useMutation({
    mutationFn: (id: number) => catalogApi.unhideDocument(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['docs-by-folder'] })
      queryClient.invalidateQueries({ queryKey: ['search', effectiveDeptId] })
      queryClient.invalidateQueries({ queryKey: ['local-dept-docs', effectiveDeptId] })
      message.success('Документ раскрыт')
    },
    onError: () => message.error('Ошибка при раскрытии документа'),
  })

  const createFolderMutation = useMutation({
    mutationFn: () => {
      if (!newFolderParentId) {
        throw new Error('parent folder is required')
      }
      return catalogApi.createFolder(effectiveDeptId, newFolderName.trim(), newFolderParentId)
    },
    onSuccess: () => {
      message.success('Папка создана')
      setCreateFolderOpen(false)
      setNewFolderName('')
      setNewFolderParentId(null)
      queryClient.invalidateQueries({ queryKey: ['folders', effectiveDeptId] })
    },
    onError: (e) => message.error(apiErr(e, 'Ошибка при создании папки')),
  })

  const moveMutation = useMutation({
    mutationFn: () => catalogApi.moveDocument(moveDoc!.id, moveFolderId!),
    onSuccess: () => {
      message.success('Документ перемещён')
      setMoveOpen(false)
      setMoveDoc(null)
      setMoveFolderId(undefined)
      queryClient.invalidateQueries({ queryKey: ['docs-by-folder'] })
      queryClient.invalidateQueries({ queryKey: ['search', effectiveDeptId] })
      queryClient.invalidateQueries({ queryKey: ['local-dept-docs', effectiveDeptId] })
    },
    onError: (e) => message.error(apiErr(e, 'Ошибка при перемещении документа')),
  })

  function handleUploadSubmit(values: { title: string; type_id: number; folder_id: number }) {
    const fileList = form.getFieldValue('file')?.fileList
    if (!fileList?.length) {
      message.error('Выберите файл')
      return
    }
    const fd = new FormData()
    fd.append('title', values.title)
    fd.append('type_id', String(values.type_id))
    fd.append('folder_id', String(values.folder_id))
    fd.append('department_id', String(effectiveDeptId))
    fd.append('file', fileList[0].originFileObj)
    uploadMutation.mutate(fd)
  }

  const mergedSearch = new Map<number, Document>()
  ;(searchResults ?? []).forEach((d) => mergedSearch.set(d.id, d))
  ;(localDeptDocs ?? [])
    .filter((d) => {
      const query = searchText.trim().toLowerCase()
      return d.title?.toLowerCase().includes(query) || d.original_name?.toLowerCase().includes(query)
    })
    .forEach((d) => mergedSearch.set(d.id, d))

  const displayDocs: Document[] = searchText ? Array.from(mergedSearch.values()) : (docs ?? [])

  const columns = [
    {
      title: 'Название',
      dataIndex: 'title',
      render: (text: string) => <Text>{text}</Text>,
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      render: (s: string) => <Tag color={STATUS_COLORS[s] ?? 'default'}>{s}</Tag>,
    },
    {
      title: 'Создан',
      dataIndex: 'created_at',
      render: (v: string) => formatRuDate(v),
    },
    {
      title: 'Действия',
      render: (_: unknown, record: Document) => (
        <Space>
          <Link to={`/documents/${record.id}`}>
            <Button size="small" type="primary">Открыть</Button>
          </Link>
          <Button
            size="small"
            onClick={() => {
              catalogApi.downloadDocument(record.id).then((blob) => {
                const url = URL.createObjectURL(blob)
                const a = document.createElement('a')
                a.href = url
                a.download = record.original_name || record.title
                a.click()
                URL.revokeObjectURL(url)
              })
            }}
          >
            Скачать
          </Button>
          <Button
            size="small"
            onClick={() => {
              setMoveDoc(record)
              setMoveFolderId(undefined)
              setMoveOpen(true)
            }}
          >
            Переместить
          </Button>
          {isAdmin && (
            <Button
              size="small"
              danger={!record.is_hidden}
              onClick={() => {
                if (record.is_hidden) {
                  unhideMutation.mutate(record.id)
                } else {
                  hideMutation.mutate(record.id)
                }
              }}
              loading={hideMutation.isPending || unhideMutation.isPending}
            >
              {record.is_hidden ? 'Скрыто' : 'Скрыть'}
            </Button>
          )}
          {!isAdmin && canAccessHeadOnly && !record.is_hidden && (
            <Button
              size="small"
              danger
              onClick={() => hideMutation.mutate(record.id)}
              loading={hideMutation.isPending}
            >
              Скрыть
            </Button>
          )}
        </Space>
      ),
    },
  ]

  const reviewColumns = [
    {
      title: 'Название',
      dataIndex: 'title',
      render: (text: string) => <Text>{text}</Text>,
    },
    {
      title: 'Тип',
      dataIndex: 'status',
      render: (s: string) => {
        const status = String(s ?? '').toUpperCase()
        if (status === 'PENDING_BOSS') return <Tag color="processing">Согласование между отделами</Tag>
        return <Tag color="warning">Визирование в отделе</Tag>
      },
    },
    {
      title: 'Поступил',
      dataIndex: 'updated_at',
      render: (_: string, record: Document) => formatRuDate(record.updated_at || record.created_at),
    },
    {
      title: 'Действия',
      render: (_: unknown, record: Document) => (
        <Space>
          <Link to={`/documents/${record.id}`}>
            <Button size="small" type="primary">Открыть</Button>
          </Link>
          <Button
            size="small"
            onClick={() => {
              catalogApi.downloadDocument(record.id).then((blob) => {
                const url = URL.createObjectURL(blob)
                const a = document.createElement('a')
                a.href = url
                a.download = record.original_name || record.title
                a.click()
                URL.revokeObjectURL(url)
              })
            }}
          >
            Скачать
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Title level={3}>Документы</Title>

      <Space style={{ marginBottom: 16 }} wrap>
        <Text type="secondary">
          {isAdmin ? 'Выберите отдел для просмотра документов' : 'Показаны документы только вашего отдела'}
        </Text>
        {isAdmin && (
          <Select
            placeholder="Отдел"
            style={{ width: 280 }}
            value={selectedDeptId ?? undefined}
            options={departments?.map((d) => ({ value: d.id, label: d.name }))}
            onChange={(v) => {
              setSelectedDeptId(v)
              setSelectedFolderId(null)
            }}
          />
        )}
        <Select
          placeholder="Выберите папку"
          style={{ width: 240 }}
          loading={foldersLoading}
          showSearch
          optionFilterProp="label"
          onChange={(v) => setSelectedFolderId(v)}
          value={selectedFolderId ?? undefined}
          options={systemFolderOptions.map((f) => ({ value: f.id, label: f.name }))}
          allowClear
        />
        <Input
          placeholder="Поиск папок"
          prefix={<FolderOpenOutlined />}
          style={{ width: 220 }}
          value={folderSearchText}
          onChange={(e) => setFolderSearchText(e.target.value)}
          allowClear
        />
        <Input
          placeholder="Поиск по части названия..."
          prefix={<SearchOutlined />}
          style={{ width: 240 }}
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          allowClear
        />
        <Button type="primary" icon={<UploadOutlined />} onClick={() => setUploadOpen(true)}>
          Загрузить документ
        </Button>
        <Button onClick={() => setCreateFolderOpen(true)} disabled={!effectiveDeptId}>
          Создать папку
        </Button>
      </Space>

      {!searchText && selectedFolderId && (
        <Space style={{ marginBottom: 12 }} wrap>
          {parentFolder && (
            <Button size="small" onClick={() => setSelectedFolderId(parentFolder.id)}>
              Вверх: {displayFolderName(parentFolder)}
            </Button>
          )}
          {childFolders.map((folder) => (
            <Button key={folder.id} size="small" onClick={() => setSelectedFolderId(folder.id)}>
              {folder.name}
            </Button>
          ))}
          {childFolders.length === 0 && (
            <Text type="secondary">Подпапок нет</Text>
          )}
        </Space>
      )}

      {canSeeReviewQueue ? (
        <Tabs
          items={[
            {
              key: 'folders',
              label: 'По папкам',
              children: (
                <Table
                  dataSource={displayDocs}
                  columns={columns}
                  rowKey="id"
                  loading={docsLoading || searching || localSearching}
                  pagination={{ pageSize: 20 }}
                />
              ),
            },
            {
              key: 'review-queue',
              label: `На визировании (${filteredReviewQueueDocs.length})`,
              children: (
                <Table
                  dataSource={filteredReviewQueueDocs}
                  columns={reviewColumns}
                  rowKey="id"
                    loading={localSearching || reviewQueueLoading}
                  pagination={{ pageSize: 20 }}
                  locale={{ emptyText: 'Нет документов, ожидающих вашего решения' }}
                />
              ),
            },
          ]}
        />
      ) : (
        <Table
          dataSource={displayDocs}
          columns={columns}
          rowKey="id"
          loading={docsLoading || searching || localSearching}
          pagination={{ pageSize: 20 }}
        />
      )}

      <Modal
        title="Загрузить документ"
        open={uploadOpen}
        onCancel={() => setUploadOpen(false)}
        footer={null}
      >
        <Form form={form} layout="vertical" onFinish={handleUploadSubmit}>
          <Form.Item label="Название" name="title" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item label="Тип документа" name="type_id" rules={[{ required: true }]}>
            <Select
              options={types?.map((t) => ({ value: t.id, label: t.name }))}
              placeholder="Выберите тип"
            />
          </Form.Item>
          <Form.Item label="Папка" name="folder_id" rules={[{ required: true }]}>
            <Select
              options={availableFoldersForActions}
              placeholder="Выберите папку"
            />
          </Form.Item>
          <Form.Item label="Файл" name="file" rules={[{ required: true }]}>
            <Upload beforeUpload={() => false} maxCount={1}>
              <Button icon={<UploadOutlined />}>Выбрать файл</Button>
            </Upload>
          </Form.Item>
          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              loading={uploadMutation.isPending}
              block
            >
              Загрузить
            </Button>
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="Создать папку"
        open={createFolderOpen}
        onCancel={() => {
          setCreateFolderOpen(false)
          setNewFolderName('')
          setNewFolderParentId(null)
        }}
        onOk={() => createFolderMutation.mutate()}
        okText="Создать"
        okButtonProps={{ disabled: !newFolderName.trim() || !effectiveDeptId || !newFolderParentId }}
        confirmLoading={createFolderMutation.isPending}
      >
        <Input
          placeholder="Название папки"
          value={newFolderName}
          onChange={(e) => setNewFolderName(e.target.value)}
          style={{ marginBottom: 12 }}
        />
        <Select
          placeholder="Родительская папка"
          style={{ width: '100%' }}
          value={newFolderParentId ?? undefined}
          onChange={(v) => setNewFolderParentId(v)}
          options={availableFoldersForActions}
        />
      </Modal>

      <Modal
        title={moveDoc ? `Переместить: ${moveDoc.title}` : 'Переместить документ'}
        open={moveOpen}
        onCancel={() => {
          setMoveOpen(false)
          setMoveDoc(null)
          setMoveFolderId(undefined)
        }}
        onOk={() => moveMutation.mutate()}
        okText="Переместить"
        okButtonProps={{ disabled: !moveFolderId }}
        confirmLoading={moveMutation.isPending}
      >
        <Select
          placeholder="Выберите папку назначения"
          style={{ width: '100%' }}
          value={moveFolderId}
          onChange={(v) => setMoveFolderId(v)}
          options={availableFoldersForActions}
        />
      </Modal>
    </div>
  )
}