import { apiRequest } from './client'

export type User = {
  id: string
  email: string
}

export type AuthResponse = {
  token: string
  user: User
}

export type MessageResponse = {
  message?: string
}

function normalizeUser(raw: unknown): User {
  if (!raw || typeof raw !== 'object') {
    throw new Error('Invalid user payload')
  }
  const record = raw as Record<string, unknown>
  const id = record.id ?? record.ID
  const email = record.email ?? record.Email
  if (id == null || typeof email !== 'string') {
    throw new Error('Invalid user payload')
  }
  return { id: String(id), email }
}

function normalizeAuthResponse(raw: unknown): AuthResponse {
  if (!raw || typeof raw !== 'object') {
    throw new Error('Invalid auth response')
  }
  const record = raw as Record<string, unknown>
  const token = record.token ?? record.Token
  if (typeof token !== 'string' || !token) {
    throw new Error('Auth response missing token')
  }
  const userRaw = record.user ?? record.User ?? record
  return { token, user: normalizeUser(userRaw) }
}

export async function register(
  email: string,
  password: string,
): Promise<AuthResponse> {
  const data = await apiRequest<unknown>('/auth/register', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
  return normalizeAuthResponse(data)
}

export async function login(
  email: string,
  password: string,
): Promise<AuthResponse> {
  const data = await apiRequest<unknown>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
  return normalizeAuthResponse(data)
}

export async function logout(): Promise<void> {
  await apiRequest<unknown>('/auth/logout', { method: 'POST' })
}

export async function forgotPassword(email: string): Promise<MessageResponse> {
  return apiRequest<MessageResponse>('/auth/forgot-password', {
    method: 'POST',
    body: JSON.stringify({ email }),
  })
}

export async function resetPassword(
  token: string,
  password: string,
): Promise<MessageResponse> {
  return apiRequest<MessageResponse>('/auth/reset-password', {
    method: 'POST',
    body: JSON.stringify({ token, password }),
  })
}

export async function changePassword(
  oldPassword: string,
  newPassword: string,
): Promise<AuthResponse | MessageResponse> {
  return apiRequest<AuthResponse | MessageResponse>('/auth/change-password', {
    method: 'POST',
    body: JSON.stringify({
      old_password: oldPassword,
      new_password: newPassword,
    }),
  })
}

export async function fetchMe(): Promise<User> {
  const data = await apiRequest<unknown>('/auth/me', { method: 'GET' })
  if (data && typeof data === 'object' && 'user' in data) {
    return normalizeUser((data as { user: unknown }).user)
  }
  return normalizeUser(data)
}
