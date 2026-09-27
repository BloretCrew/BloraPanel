import { reactive } from 'vue'
export interface User { userId: string; name: string; admin: boolean; disabled: boolean; revision: number }
export interface Node { nodeId: string; name: string; group?:string; tags?:string[]; platform: string; state: string; generation: number; lastSeen: string; capabilities: Record<string,string>; revision: number;maintenance?:boolean;quota?:number;configRevision?:number }
export interface Audit { sequence:number; actorId:string; nodeId?:string; resourceKey:string; action:string; requestId?:string; result:string; recordedAt:string }
export interface Instance { instanceId: string; nodeId: string; nodeName?:string;nodeState?:string; name: string; group: string; tags: string[]; config: Record<string, unknown>; state: string; runId?: string; revision: number;configRevision?:number }
export interface Task { taskId: string; actorId?:string; updatedAt?:string; retryOf?:string; requestId: string; action: string; resource: { kind: string; id: string; nodeId?: string }; state: string; phase: string; error?: string; revision: number; createdAt: string;cancellationRequested?:boolean;cancellationRequestId?:string;dispatchedAt?:string;result?:Record<string,unknown> }
export const session = reactive<{ user?: User; csrfToken: string; ready: boolean; error: string }>({ csrfToken: '', ready: false, error: '' })
export class APIError extends Error { constructor(message: string, readonly code: string, readonly status: number) { super(message) } }
export async function api<T>(path: string, options: RequestInit = {}): Promise<T> {
  const requestUserId=session.user?.userId
  const headers = new Headers(options.headers)
  if (options.body) headers.set('Content-Type', 'application/json')
  if (options.method && !['GET','HEAD'].includes(options.method)) {
    headers.set('X-CSRF-Token', session.csrfToken)
    if (!headers.has('Idempotency-Key')) headers.set('Idempotency-Key', crypto.randomUUID())
  }
  const response = await fetch(`/api/v1${path}`, { ...options, headers, credentials: 'same-origin' })
  const value = response.status === 204 ? {} : await response.json().catch(() => ({}))
  if (!response.ok) {
    if (response.status === 401 && path !== '/login' && session.user?.userId===requestUserId) session.user = undefined
    throw new APIError(value.error?.message || `请求失败 (${response.status})`, value.error?.code || 'HTTP_ERROR', response.status)
  }
  return value as T
}
export async function checkSession() {
  try { Object.assign(session, await api('/session')) }
  catch (error) { if (!(error instanceof APIError && error.status === 401)) session.error = String(error) }
  finally { session.ready = true }
}
export async function login(name: string, password: string) { Object.assign(session, await api('/login', { method: 'POST', body: JSON.stringify({ name, password }) }));session.error='' }
