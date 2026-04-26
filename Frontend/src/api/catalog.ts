import api from './client'
import type { Document, DocumentHistory, DocumentType, Folder } from '../types'

export const catalogApi = {
  listTypes: () =>
    api.get<DocumentType[]>('/api/catalog/types').then(r => r.data),

  createType: (name: string) =>
    api.post<{ id: number }>('/api/catalog/types', { name }).then(r => r.data),

  deleteType: (id: number) =>
    api.delete(`/api/catalog/types/${id}`),

  listFolders: (dept_id: number) =>
    api.get<Folder[]>(`/api/catalog/departments/${dept_id}/folders`).then(r => r.data),

  createFolder: (dept_id: number, name: string, parent_id?: number) =>
    api.post<{ id: number }>(`/api/catalog/departments/${dept_id}/folders`, { name, parent_id }).then(r => r.data),

  listDocumentsByFolder: (folder_id: number) =>
    api.get<Document[]>(`/api/catalog/folders/${folder_id}/documents`).then(r => r.data),

  uploadDocument: (formData: FormData) =>
    api.post<{ id: number }>('/api/catalog/documents', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }).then(r => r.data),

  getDocument: (id: number) =>
    api.get<Document>(`/api/catalog/documents/${id}`).then(r => r.data),

  getDocumentHistory: (id: number) =>
    api.get<DocumentHistory[]>(`/api/catalog/documents/${id}/history`).then(r => r.data),

  downloadDocument: (id: number) =>
    api.get(`/api/catalog/documents/${id}/download`, { responseType: 'blob' }).then(r => r.data),

  hideDocument: (id: number) =>
    api.post(`/api/catalog/documents/${id}/hide`),

  unhideDocument: (id: number) =>
    api.post(`/api/catalog/documents/${id}/unhide`),

  searchDocuments: (dept_id: number, q: string) =>
    api.get<Document[]>(`/api/catalog/departments/${dept_id}/search`, { params: { q } }).then(r => r.data),
}
