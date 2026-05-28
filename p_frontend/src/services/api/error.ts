import axios from 'axios'

export type ApiError = {
  status?: number
  code?: string | number
  message: string
}

function readMessage(data: unknown): string | undefined {
  if (data && typeof data === 'object') {
    const objectData = data as Record<string, unknown>
    const message = objectData.message

    if (typeof message === 'string' && message.trim().length > 0) {
      return message
    }
  }

  return undefined
}

function readBusinessCode(data: unknown): number | undefined {
  if (data && typeof data === 'object') {
    const objectData = data as Record<string, unknown>
    const code = objectData.code
    if (typeof code === 'number') {
      return code
    }
  }
  return undefined
}

export function normalizeApiError(error: unknown): ApiError {
  if (
    error &&
    typeof error === 'object' &&
    'response' in error &&
    error.response &&
    typeof error.response === 'object'
  ) {
    const response = error.response as { status?: number; data?: unknown }

    const businessCode = readBusinessCode(response.data)
    return {
      status: businessCode ?? response.status,
      code: businessCode,
      message: readMessage(response.data) ?? '请求失败，请稍后重试。',
    }
  }

  if (axios.isAxiosError(error)) {
    return {
      status: error.response?.status,
      code: error.code,
      message:
        readMessage(error.response?.data) ??
        error.message ??
        '请求失败，请稍后重试。',
    }
  }

  if (error instanceof Error) {
    return {
      message: error.message || '请求失败，请稍后重试。',
    }
  }

  return {
    message: '请求失败，请稍后重试。',
  }
}
