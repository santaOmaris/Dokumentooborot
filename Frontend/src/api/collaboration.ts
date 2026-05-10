import api from './client'
import type { AuditLog, Message } from '../types'

export const collaborationApi = {
  listMessages: (doc_id: number) =>
    api.get<Message[]>(`/api/collaboration/documents/${doc_id}/messages`).then(r => r.data),

  sendMessage: (doc_id: number, content: string) =>
    api.post(`/api/collaboration/documents/${doc_id}/messages`, { content }),

  getDeptAudit: (dept_id: number) =>
    api.get<AuditLog[]>(`/api/collaboration/departments/${dept_id}/audit`).then(r => r.data),

  getDocAudit: (doc_id: number) =>
    api.get<AuditLog[]>(`/api/collaboration/documents/${doc_id}/audit`).then(r => r.data),
}
