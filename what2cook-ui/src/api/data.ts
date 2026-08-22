import { ApiError, getToken } from './client'

export type ImportResult = {
  inventories: number
  items: number
  skipped: number
}

export type ExportFormat = 'csv' | 'xlsx'

function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

export async function exportUserData(format: ExportFormat): Promise<void> {
  const token = getToken()
  const headers = new Headers()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  let response: Response
  try {
    response = await fetch(`/api/v1/data/export?format=${format}`, { headers })
  } catch (err) {
    console.error('Export request failed', err)
    throw new ApiError('Unable to reach the server. Please try again.', 0, null)
  }

  if (!response.ok) {
    let body: unknown = null
    try {
      body = await response.json()
    } catch {
      body = null
    }
    const message =
      body && typeof body === 'object' && 'error' in body && typeof (body as { error: unknown }).error === 'string'
        ? (body as { error: string }).error
        : response.statusText || 'Export failed'
    throw new ApiError(message, response.status, body)
  }

  const blob = await response.blob()
  const filename =
    format === 'xlsx' ? 'what2cook-export.xlsx' : 'what2cook-export.csv'
  downloadBlob(blob, filename)
}

export async function importUserData(file: File): Promise<ImportResult> {
  const token = getToken()
  const headers = new Headers()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const form = new FormData()
  form.append('file', file)

  const ext = file.name.toLowerCase()
  if (ext.endsWith('.xlsx') || ext.endsWith('.xls')) {
    form.append('format', 'xlsx')
  } else {
    form.append('format', 'csv')
  }

  let response: Response
  try {
    response = await fetch('/api/v1/data/import', {
      method: 'POST',
      headers,
      body: form,
    })
  } catch (err) {
    console.error('Import request failed', err)
    throw new ApiError('Unable to reach the server. Please try again.', 0, null)
  }

  let body: unknown = null
  try {
    body = await response.json()
  } catch {
    body = null
  }

  if (!response.ok) {
    const message =
      body && typeof body === 'object' && 'error' in body && typeof (body as { error: unknown }).error === 'string'
        ? (body as { error: string }).error
        : response.statusText || 'Import failed'
    throw new ApiError(message, response.status, body)
  }

  const result = body as ImportResult
  if (
    !result ||
    typeof result.inventories !== 'number' ||
    typeof result.items !== 'number'
  ) {
    throw new Error('Unexpected import response from server')
  }

  return {
    inventories: result.inventories,
    items: result.items,
    skipped: typeof result.skipped === 'number' ? result.skipped : 0,
  }
}
