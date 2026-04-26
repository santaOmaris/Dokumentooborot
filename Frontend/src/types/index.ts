// ─── IAM ─────────────────────────────────────────────────────────────────────

export interface CurrentUser {
  user_id: number
  login: string
  system_role: 'USER' | 'ADMIN'
  is_head: boolean
  department_id: string
}

export interface User {
  id: number
  login: string
  email: string
  full_name: string
  is_head: boolean
  system_role: string
  department_id: number | null
  department_name?: string
}

export interface Department {
  id: number
  name: string
  parent_id: number | null
}

// ─── Catalog ──────────────────────────────────────────────────────────────────

export interface DocumentType {
  id: number
  name: string
}

export interface Folder {
  id: number
  name: string
  department_id: number
  parent_id: number | null
  is_system?: boolean
}

export interface Document {
  id: number
  title: string
  folder_id: number
  type_id: number
  owner_id: number
  status: string
  created_at: string
  updated_at: string
}

export interface DocumentHistory {
  id: number
  document_id: number
  changed_by: number
  change_type: string
  description: string
  created_at: string
}

// ─── Collaboration ────────────────────────────────────────────────────────────

export interface Message {
  id: number
  document_id: number
  sender_id: number
  sender_login: string
  content: string
  created_at: string
}

export interface AuditLog {
  id: number
  document_id: number | null
  department_id: string
  actor_login: string
  action: string
  details: string
  created_at: string
}

// ─── Orchestrator ─────────────────────────────────────────────────────────────

export interface DocumentStatus {
  document_id: number
  status: string
  assignee_id: number | null
}

export interface WorkflowEvent {
  id: number
  document_id: number
  from_state: string
  to_state: string
  triggered_by: string
  note: string
  created_at: string
}
