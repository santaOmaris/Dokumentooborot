import { useMemo, useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Table,
  Button,
  Space,
  Typography,
  Select,
  Tag,
  Upload,
  Modal,
  Form,
  Input,
  message,
} from 'antd'
import { UploadOutlined, SearchOutlined } from '@ant-design/icons'
import { Link } from 'react-router-dom'
import { useMe } from '../hooks/useMe'
import { catalogApi } from '../api/catalog'
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

type FolderGroup = {
  key: string
  name: string
  folderIds: number[]
  headOnly: boolean
}

function normalizeFolderName(name: string): string {
  const trimmed = name.trim()
  const translated = SYSTEM_FOLDER_LABELS[trimmed.toLowerCase()]
  return translated ?? trimmed
}

function buildFolderGroups(folders: Folder[]): FolderGroup[] {
  const map = new Map<string, FolderGroup>()
  folders.forEach((folder) => {
    const canonicalName = normalizeFolderName(folder.name)
    const key = canonicalName.toLowerCase()
    const existing = map.get(key)
    if (existing) {
      existing.folderIds.push(folder.id)
      existing.headOnly = existing.headOnly || (Boolean(folder.is_system) && folder.name === 'head_only')
      return
    }
    map.set(key, {
      key,
      name: canonicalName,
      folderIds: [folder.id],
      headOnly: Boolean(folder.is_system) && folder.name === 'head_only',
    })
  })
  return Array.from(map.values()).sort((a, b) => a.name.localeCompare(b.name, 'ru-RU'))
}

function getUploadFolderId(group: FolderGroup, allFolders: Folder[]): number {
  const groupFolders = allFolders.filter((f) => group.folderIds.includes(f.id))
  const systemFolder = groupFolders.find((f) => f.is_system)
  return systemFolder?.id ?? group.folderIds[0]
}

export default function DocumentsPage() {
  const { data: me } = useMe()
  const myDeptId = me ? Number(me.department_id) : 0
  const role = String(me?.system_role ?? '').toUpperCase()
  const canAccessHeadOnly = Boolean(me?.is_head || role === 'ADMIN')

  const [selectedFolderKey, setSelectedFolderKey] = useState<string | null>(null)
  const [searchText, setSearchText] = useState('')
  const [uploadOpen, setUploadOpen] = useState(false)
  const [form] = Form.useForm()
  const queryClient = useQueryClient()

  const { data: folders, isLoading: foldersLoading } = useQuery({
    queryKey: ['folders', myDeptId],
    queryFn: () => catalogApi.listFolders(myDeptId),
    enabled: !!myDeptId,
  })

  const folderGroups = useMemo(() => buildFolderGroups(folders ?? []), [folders])
  const visibleFolderGroups = useMemo(
    () => folderGroups.filter((g) => !g.headOnly || canAccessHeadOnly),
    [folderGroups, canAccessHeadOnly]
  )
  const selectedFolderGroup = useMemo(
    () => visibleFolderGroups.find((g) => g.key === selectedFolderKey) ?? null,
    [visibleFolderGroups, selectedFolderKey]
  )

  const { data: types } = useQuery({
    queryKey: ['types'],
    queryFn: catalogApi.listTypes,
  })

  const { data: docs, isLoading: docsLoading } = useQuery({
    queryKey: ['docs-by-folder-group', selectedFolderKey, selectedFolderGroup?.folderIds.join(',') ?? ''],
    queryFn: async () => {
      if (!selectedFolderGroup) return []
      const all = await Promise.all(
        selectedFolderGroup.folderIds.map((folderId) =>
          catalogApi.listDocumentsByFolder(folderId).catch(() => [])
        )
      )
      const unique = new Map<number, Document>()
      all.flat().forEach((d) => unique.set(d.id, d))
      return Array.from(unique.values())
    },
    enabled: !!selectedFolderGroup,
  })

  const { data: searchResults, isFetching: searching } = useQuery({
    queryKey: ['search', myDeptId, searchText],
    queryFn: () => catalogApi.searchDocuments(myDeptId, searchText),
    enabled: !!searchText && !!myDeptId,
  })

  const { data: localDeptDocs, isFetching: localSearching } = useQuery({
    queryKey: ['local-dept-docs', myDeptId],
    queryFn: async () => {
      const deptFolders = await catalogApi.listFolders(myDeptId)
      const all = await Promise.all(deptFolders.map((f) => catalogApi.listDocumentsByFolder(f.id).catch(() => [])))
      const unique = new Map<number, Document>()
      all.flat().forEach((d) => unique.set(d.id, d))
      return Array.from(unique.values())
    },
    enabled: !!searchText && !!myDeptId,
  })

  const uploadMutation = useMutation({
    mutationFn: (fd: FormData) => catalogApi.uploadDocument(fd),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['docs-by-folder-group'] })
      queryClient.invalidateQueries({ queryKey: ['search', myDeptId] })
      queryClient.invalidateQueries({ queryKey: ['local-dept-docs', myDeptId] })
      message.success('Документ загружен')
      setUploadOpen(false)
      form.resetFields()
    },
    onError: () => message.error('Ошибка при загрузке'),
  })

  const hideMutation = useMutation({
    mutationFn: (id: number) => catalogApi.hideDocument(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['docs-by-folder-group'] })
      queryClient.invalidateQueries({ queryKey: ['search', myDeptId] })
      queryClient.invalidateQueries({ queryKey: ['local-dept-docs', myDeptId] })
      message.success('Документ скрыт')
    },
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
    fd.append('department_id', String(myDeptId))
    fd.append('file', fileList[0].originFileObj)
    uploadMutation.mutate(fd)
  }

  const mergedSearch = new Map<number, Document>()
  ;(searchResults ?? []).forEach((d) => mergedSearch.set(d.id, d))
  ;(localDeptDocs ?? [])
    .filter((d) => d.title?.toLowerCase().includes(searchText.trim().toLowerCase()))
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
      render: (v: string) => new Date(v).toLocaleDateString('ru-RU'),
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
                a.download = record.title
                a.click()
                URL.revokeObjectURL(url)
              })
            }}
          >
            Скачать
          </Button>
          {record.status !== 'HIDDEN' && (
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

  return (
    <div>
      <Title level={3}>Документы</Title>

      <Space style={{ marginBottom: 16 }} wrap>
        <Text type="secondary">Показаны документы только вашего отдела</Text>
        <Select
          placeholder="Выберите папку"
          style={{ width: 240 }}
          loading={foldersLoading}
          onChange={(v) => setSelectedFolderKey(v)}
          value={selectedFolderKey ?? undefined}
          options={visibleFolderGroups.map((g) => ({ value: g.key, label: g.name }))}
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
      </Space>

      <Table
        dataSource={displayDocs}
        columns={columns}
        rowKey="id"
        loading={docsLoading || searching || localSearching}
        pagination={{ pageSize: 20 }}
      />

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
                options={visibleFolderGroups.map((g) => ({ value: getUploadFolderId(g, folders ?? []), label: g.name }))}
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
    </div>
  )
}
