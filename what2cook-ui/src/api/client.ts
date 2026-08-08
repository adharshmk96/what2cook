const TOKEN_KEY = 'what2cook_token'

export function getToken(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY)
  } catch {
    return null
  }
}

export function setToken(token: string | null): void {
  try {
    if (token) {
      localStorage.setItem(TOKEN_KEY, token)
    } else {
      localStorage.removeItem(TOKEN_KEY)
    }
  } catch (err) {
    console.warn('Failed to persist auth token', err)
  }
}

export class ApiError extends Error {
  readonly status: number
  readonly body: unknown

  constructor(message: string, status: number, body: unknown) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.body = body
  }
}

function errorMessage(body: unknown, fallback: string): string {
  if (body && typeof body === 'object') {
    const record = body as Record<string, unknown>
    for (const key of ['error', 'message', 'msg'] as const) {
      const value = record[key]
      if (typeof value === 'string' && value.trim()) {
        return value
      }
    }
  }
  return fallback
}

export async function apiRequest<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const headers = new Headers(options.headers)
  if (!headers.has('Content-Type') && options.body != null) {
    headers.set('Content-Type', 'application/json')
  }

  const token = getToken()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  let response: Response
  try {
    response = await fetch(`/api/v1${path}`, {
      ...options,
      headers,
    })
  } catch (err) {
    console.error('Network request failed', path, err)
    throw new ApiError('Unable to reach the server. Please try again.', 0, null)
  }

  const contentType = response.headers.get('content-type') ?? ''
  const isJson = contentType.includes('application/json')
  let body: unknown = null

  if (response.status !== 204) {
    try {
      body = isJson ? await response.json() : await response.text()
    } catch {
      body = null
    }
  }

  if (!response.ok) {
    const message = errorMessage(
      body,
      response.statusText || `Request failed (${response.status})`,
    )
    console.warn('API error', path, response.status, message)
    throw new ApiError(message, response.status, body)
  }

  return body as T
}
