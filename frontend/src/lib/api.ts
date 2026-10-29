export type User = {
  id: string
  username: string
  role: 'admin' | 'user'
  hasPassword: boolean
  hasOIDC: boolean
  authMethod?: 'password' | 'oidc'
}
export type AuthConfig = { passwordLoginEnabled: boolean; oidcEnabled: boolean; oidcButtonTitle: string; bootstrap: boolean }
export type Check = { id?: number; is_up: boolean; status_code?: number | null; response_time_ms?: number; checked_at: string }
export type Monitor = {
  id: number
  owner_id: string
  type: 'http' | 'tcp'
  name: string
  url: string
  interval_seconds: number
  fail_threshold: number
  paused: boolean
  created_at: string
  notifier_ids: number[]
  status_code: number | null
  response_time_ms: number | null
  is_up: boolean | null
  checked_at: string | null
  uptime_24h: number | null
  recent_checks: Check[]
}
export type MonitorInput = Pick<Monitor, 'name' | 'type' | 'url' | 'interval_seconds' | 'fail_threshold' | 'notifier_ids'>
export type Notifier = {
  id: number
  owner_id: string
  name: string
  type: 'telegram' | 'pushover' | 'redis'
  config_json: string
  created_at: string
}
export type NotifierInput = Pick<Notifier, 'name' | 'type' | 'config_json'>
export type UserInput = { username: string; role: User['role']; password?: string; newPassword?: string; resetOIDC?: boolean }
export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
  ) {
    super(message)
  }
}

async function request<T>(path: string, method = 'GET', body?: unknown, signal?: AbortSignal, anonymous = false): Promise<T> {
  const response = await fetch('/api' + path, {
    method,
    credentials: 'same-origin',
    signal,
    headers: { 'Content-Type': 'application/json', 'X-UptimeLit-Request': '1' },
    ...(body === undefined ? {} : { body: JSON.stringify(body) }),
  })
  if (response.status === 204) return undefined as T
  const data = await response.json().catch(() => ({ error: 'The server returned an invalid response' }))
  if (!response.ok) {
    if (response.status === 401 && !anonymous) window.dispatchEvent(new Event('uptimelit:unauthorized'))
    throw new ApiError(data.error || `Request failed (${response.status})`, response.status)
  }
  return data as T
}
export const api = {
  authConfig: (signal?: AbortSignal) => request<AuthConfig>('/auth/config', 'GET', undefined, signal, true),
  me: () => request<User>('/auth/me', 'GET', undefined, undefined, true),
  login: (username: string, password: string) => request<{ user: User }>('/auth/login', 'POST', { username, password }, undefined, true),
  logout: () => request('/auth/logout', 'POST'),
  profile: (body: { username: string; currentPassword?: string; newPassword?: string }) => request('/auth/profile', 'PUT', body),
  status: () => request<{ version: string }>('/status'),
  monitors: (signal?: AbortSignal) => request<Monitor[]>('/endpoints', 'GET', undefined, signal),
  saveMonitor: (body: MonitorInput, id?: number) => request<Monitor>(id ? `/endpoints/${id}` : '/endpoints', id ? 'PUT' : 'POST', body),
  deleteMonitor: (id: number) => request(`/endpoints/${id}`, 'DELETE'),
  pauseMonitor: (id: number) => request(`/endpoints/${id}/pause`, 'POST'),
  checks: (id: number, signal?: AbortSignal) => request<Check[] | null>(`/endpoints/${id}/checks?limit=100`, 'GET', undefined, signal),
  notifiers: (signal?: AbortSignal) => request<Notifier[]>('/notifiers', 'GET', undefined, signal),
  saveNotifier: (body: NotifierInput, id?: number) => request<Notifier>(id ? `/notifiers/${id}` : '/notifiers', id ? 'PUT' : 'POST', body),
  deleteNotifier: (id: number) => request(`/notifiers/${id}`, 'DELETE'),
  testNotifier: (id: number) => request(`/notifiers/${id}/test`, 'POST'),
  users: (signal?: AbortSignal) => request<User[]>('/admin/users', 'GET', undefined, signal),
  saveUser: (body: UserInput, id?: string) =>
    request<User>(id ? `/admin/users/${encodeURIComponent(id)}` : '/admin/users', id ? 'PUT' : 'POST', body),
  deleteUser: (id: string) => request(`/admin/users/${encodeURIComponent(id)}`, 'DELETE'),
}
