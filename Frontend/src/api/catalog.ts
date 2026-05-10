import api from './client'
import type { Document, DocumentHistory, DocumentType, Folder } from '../types'

function pick<T>(src: Record<string, unknown>, ...keys: string[]): T | undefined {
  for (const key of keys) {
    if (key in src) return src[key] as T
  }
  return undefined
}

function num(value: unknown, fallback = 0): number {
  if (typeof value === 'number') return value
  if (typeof value === 'string' && value.trim() !== '') {
    const parsed = Number(value)
    return Number.isFinite(parsed) ? parsed : fallback
  }
  if (value && typeof value === 'object') {
    const boxed = value as { Int32?: number; Valid?: boolean }
    if (boxed.Valid && typeof boxed.Int32 === 'number') return boxed.Int32
  }
  return fallback
}

function str(value: unknown, fallback = ''): string {
  return typeof value === 'string' ? value : fallback
}

function timeStr(value: unknown, fallback = ''): string {
  if (typeof value === 'string') return value
  if (value && typeof value === 'object') {
    const boxed = value as { Time?: unknown; Valid?: unknown }
    if (boxed.Valid === true && typeof boxed.Time === 'string') {
      return boxed.Time
    }
  }
  return fallback
}

function normalizeDocument(raw: unknown): Document {
  const d = (raw ?? {}) as Record<string, unknown>
  return {
    id: num(pick(d, 'id', 'ID')),
    title: str(pick(d, 'title', 'Title')),
    original_name: str(pick(d, 'original_name', 'OriginalName')),
    folder_id: num(pick(d, 'folder_id', 'FolderID')),
    type_id: (() => {
      const v = pick(d, 'type_id', 'TypeID')
      if (v == null) return null
      const n = num(v, NaN)
      return Number.isFinite(n) ? n : null
    })(),
    owner_id: num(pick(d, 'owner_id', 'created_by', 'OwnerID', 'CreatedBy')),
    created_by: num(pick(d, 'created_by', 'CreatedBy')),
    assignee_id: (() => {
      const v = pick(d, 'assignee_id', 'AssigneeID')
      if (v == null) return null
      const n = num(v, NaN)
      return Number.isFinite(n) ? n : null
    })(),
    department_id: num(pick(d, 'department_id', 'DepartmentID')),
    is_hidden: Boolean(pick(d, 'is_hidden', 'IsHidden')),
    status: str(pick(d, 'status', 'Status'), 'DRAFT'),
    created_at: timeStr(pick(d, 'created_at', 'CreatedAt')),
    updated_at: timeStr(pick(d, 'updated_at', 'UpdatedAt')),
  }
}

function normalizeFolder(raw: unknown): Folder {
  const f = (raw ?? {}) as Record<string, unknown>
  return {
    id: num(pick(f, 'id', 'ID')),
    name: str(pick(f, 'name', 'Name')),
    department_id: num(pick(f, 'department_id', 'DepartmentID')),
    parent_id: (() => {
      const v = pick(f, 'parent_id', 'ParentID')
      if (v == null) return null
      const n = num(v, NaN)
      return Number.isFinite(n) ? n : null
    })(),
    is_system: Boolean(pick(f, 'is_system', 'IsSystem')),
  }
}

function normalizeDocumentHistory(raw: unknown): DocumentHistory {
  const h = (raw ?? {}) as Record<string, unknown>
  return {
    id: num(pick(h, 'id', 'ID')),
    document_id: num(pick(h, 'document_id', 'DocumentID')),
    changed_by: num(pick(h, 'changed_by', 'ChangedBy')),
    actor_login: str(pick(h, 'actor_login', 'ActorLogin')),
    change_type: str(pick(h, 'change_type', 'action', 'ChangeType', 'Action')),
    description: str(pick(h, 'description', 'details', 'Description', 'Details')),
    created_at: timeStr(pick(h, 'created_at', 'CreatedAt')),
  }
}

export const catalogApi = {
  listTypes: () =>
    api.get<DocumentType[]>('/api/catalog/types').then(r => r.data),

  createType: (name: string) =>
    api.post<{ id: number }>('/api/catalog/types', { name }).then(r => r.data),

  deleteType: (id: number) =>
    api.delete(`/api/catalog/types/${id}`),

  listFolders: (dept_id: number) =>
    api.get<unknown[]>(`/api/catalog/departments/${dept_id}/folders`).then(r => (r.data ?? []).map(normalizeFolder)),

  createFolder: (dept_id: number, name: string, parent_id?: number) =>
    api.post<{ id: number }>(`/api/catalog/departments/${dept_id}/folders`, { name, parent_id }).then(r => r.data),

  listDocumentsByFolder: (folder_id: number) =>
    api.get<unknown[]>(`/api/catalog/folders/${folder_id}/documents`).then(r => (r.data ?? []).map(normalizeDocument)),

  uploadDocument: (formData: FormData) =>
    api.post<{ id: number }>('/api/catalog/documents', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }).then(r => r.data),

  getDocument: (id: number) =>
    api.get<unknown>(`/api/catalog/documents/${id}`).then(r => normalizeDocument(r.data)),

  getDocumentHistory: (id: number) =>
    api.get<unknown[]>(`/api/catalog/documents/${id}/history`).then(r => (r.data ?? []).map(normalizeDocumentHistory)),

  downloadDocument: (id: number) =>
    api.get(`/api/catalog/documents/${id}/download`, { responseType: 'blob' }).then(r => r.data),

  hideDocument: (id: number) =>
    api.post(`/api/catalog/documents/${id}/hide`),

  moveDocument: (id: number, folder_id: number) =>
    api.patch(`/api/catalog/documents/${id}/move`, { folder_id }),

  unhideDocument: (id: number) =>
    api.post(`/api/catalog/documents/${id}/unhide`),

  searchDocuments: (dept_id: number, q: string) =>
    api.get<unknown[]>(`/api/catalog/departments/${dept_id}/search`, { params: { q } }).then(r => (r.data ?? []).map(normalizeDocument)),
}
