import api from './client'
import type { CurrentUser, User, Department } from '../types'

function normalizeBoolean(value: unknown): boolean {
  if (typeof value === 'boolean') return value
  if (typeof value === 'number') return value !== 0
  if (typeof value === 'string') {
    const normalized = value.trim().toLowerCase()
    return normalized === 'true' || normalized === '1' || normalized === 't' || normalized === 'yes'
  }
  return false
}

function normalizeDepartmentId(value: unknown): number | null {
  if (typeof value === 'number') return value
  if (typeof value === 'string' && value.trim() !== '') {
    const parsed = Number(value)
    return Number.isFinite(parsed) ? parsed : null
  }
  if (!value || typeof value !== 'object') return null
  const maybe = value as { Int32?: number; Int64?: number; Valid?: boolean }
  if (maybe.Valid && typeof maybe.Int32 === 'number') return maybe.Int32
  if (maybe.Valid && typeof maybe.Int64 === 'number') return maybe.Int64
  return null
}

function normalizeNumber(value: unknown): number {
  if (typeof value === 'number') return value
  if (typeof value === 'string' && value.trim() !== '') {
    const parsed = Number(value)
    return Number.isFinite(parsed) ? parsed : 0
  }
  return 0
}

function normalizeUser(raw: unknown, fallbackDeptId?: number): User {
  const u = (raw ?? {}) as {
    id?: number
    login?: string
    email?: string
    full_name?: string
    is_head?: boolean
    system_role?: string
    department_id?: unknown
  }

  return {
    id: Number(u.id ?? 0),
    login: String(u.login ?? ''),
    email: String(u.email ?? ''),
    full_name: String(u.full_name ?? ''),
    is_head: Boolean(u.is_head),
    system_role: String(u.system_role ?? 'USER'),
    department_id: normalizeDepartmentId(u.department_id) ?? fallbackDeptId ?? null,
  }
}

function normalizeCurrentUser(raw: unknown): CurrentUser {
  const u = (raw ?? {}) as Record<string, unknown>

  const role = String((u.system_role ?? u.systemRole ?? 'USER')).toUpperCase()
  const department = normalizeDepartmentId(u.department_id ?? u.departmentId)

  return {
    user_id: normalizeNumber(u.user_id ?? u.userId),
    login: String(u.login ?? ''),
    system_role: (role === 'ADMIN' ? 'ADMIN' : 'USER') as CurrentUser['system_role'],
    is_head: normalizeBoolean(u.is_head ?? u.isHead),
    department_id: department != null ? String(department) : '',
  }
}

function normalizeDepartment(raw: unknown): Department {
  const d = (raw ?? {}) as Record<string, unknown>
  return {
    id: normalizeNumber(d.id ?? d.ID),
    name: String(d.name ?? d.Name ?? ''),
    parent_id: normalizeDepartmentId(d.parent_id ?? d.parentId ?? d.ParentID),
  }
}

export const iamApi = {
  login: (login: string, password: string) =>
    api.post('/api/iam/auth/login', { login, password }),

  logout: () => api.post('/api/iam/auth/logout'),

  me: () => api.get<unknown>('/api/iam/auth/me').then(r => normalizeCurrentUser(r.data)),

  register: (data: {
    login: string
    email: string
    password: string
    full_name: string
    is_head: boolean
    system_role: string
    department_id?: number
  }) => api.post<{ user_id: number }>('/api/iam/auth/register', data).then(r => r.data),

  listUsers: () =>
    api.get<unknown[]>('/api/iam/users').then(r => (r.data ?? []).map((u) => normalizeUser(u))),

  listDepartments: () =>
    api.get<unknown[]>('/api/iam/departments').then(r => (r.data ?? []).map((d) => normalizeDepartment(d))),

  listDeptUsers: (id: number) =>
    api.get<unknown[]>(`/api/iam/departments/${id}/users`).then(r => (r.data ?? []).map((u) => normalizeUser(u, id))),

  fireUser: (user_id: number) =>
    api.post('/api/iam/users/fire', { user_id }),

  moveUser: (user_id: number, department_id: number) =>
    api.post('/api/iam/users/move', { user_id, department_id }),

  promoteUser: (id: number) =>
    api.post(`/api/iam/users/${id}/promote`),

  demoteUser: (id: number) =>
    api.post(`/api/iam/users/${id}/demote`),

  createDepartment: (name: string, parent_id?: number) =>
    api.post<{ id: number }>('/api/iam/departments', { name, parent_id }),

  setDeptParent: (id: number, parent_id: number | null) =>
    api.patch(`/api/iam/departments/${id}/parent`, { parent_id }),

  deleteDepartment: (id: number) =>
    api.delete(`/api/iam/departments/${id}`),
}
