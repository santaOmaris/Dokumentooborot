import api from './client'
import type { DocumentStatus, WorkflowEvent } from '../types'

function toNumber(value: unknown, fallback = 0): number {
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

function toString(value: unknown, fallback = ''): string {
  return typeof value === 'string' ? value : fallback
}

function normalizeStatus(raw: unknown): DocumentStatus {
  const s = (raw ?? {}) as Record<string, unknown>
  const approverRaw = s.approver_id ?? s.ApproverID ?? s.assignee_id ?? s.AssigneeID
  const approver = approverRaw == null ? null : toNumber(approverRaw, NaN)

  return {
    document_id: toNumber(s.document_id ?? s.DocumentID),
    status: toString(s.status ?? s.Status ?? s.state ?? s.State),
    assignee_id: Number.isFinite(approver) ? approver : null,
  }
}

function normalizeEvent(raw: unknown): WorkflowEvent {
  const e = (raw ?? {}) as Record<string, unknown>
  return {
    id: toNumber(e.id ?? e.ID),
    document_id: toNumber(e.document_id ?? e.DocumentID),
    from_state: toString(e.from_state ?? e.FromState),
    to_state: toString(e.to_state ?? e.ToState),
    triggered_by: toString(e.triggered_by ?? e.TriggeredBy ?? e.actor_login ?? e.ActorLogin),
    note: toString(e.note ?? e.Note ?? e.revision_note ?? e.RevisionNote),
    created_at: toString(e.created_at ?? e.CreatedAt),
  }
}

export const orchestratorApi = {
  getStatus: (doc_id: number) =>
    api.get<unknown>(`/api/orchestrator/documents/${doc_id}/status`).then(r => normalizeStatus(r.data)),

  getHistory: (doc_id: number) =>
    api.get<unknown[]>(`/api/orchestrator/documents/${doc_id}/history`).then(r => (r.data ?? []).map(normalizeEvent)),

  sendForVisa: (doc_id: number, payload?: { note?: string; approver_id?: number }) =>
    api.post(`/api/orchestrator/documents/${doc_id}/send-for-visa`, {
      note: payload?.note ?? '',
      ...(payload?.approver_id ? { approver_id: payload.approver_id } : {}),
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
