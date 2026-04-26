import api from './client'
import type { DocumentStatus, WorkflowEvent } from '../types'

export const orchestratorApi = {
  getStatus: (doc_id: number) =>
    api.get<DocumentStatus>(`/api/orchestrator/documents/${doc_id}/status`).then(r => r.data),

  getHistory: (doc_id: number) =>
    api.get<WorkflowEvent[]>(`/api/orchestrator/documents/${doc_id}/history`).then(r => r.data),

  sendForVisa: (doc_id: number, payload?: { note?: string }) =>
    api.post(`/api/orchestrator/documents/${doc_id}/send-for-visa`, {
      note: payload?.note ?? '',
    }),

  approve: (doc_id: number) =>
    api.post(`/api/orchestrator/documents/${doc_id}/approve`),

  reject: (doc_id: number, revision_note: string) =>
    api.post(`/api/orchestrator/documents/${doc_id}/reject`, { revision_note }),

  delegate: (doc_id: number, assignee_id: number) =>
    api.post(`/api/orchestrator/documents/${doc_id}/delegate`, { assignee_id }),

  requestApproval: (doc_id: number, payload?: { question?: string; target_department_id?: number }) =>
    api.post(`/api/orchestrator/documents/${doc_id}/request-approval`, {
      question: payload?.question ?? '',
      ...(payload?.target_department_id ? { target_department_id: payload.target_department_id } : {}),
    }),
}
