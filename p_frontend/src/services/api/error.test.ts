import { describe, expect, it } from 'vitest'
import { normalizeApiError } from './error'

describe('normalizeApiError', () => {
  it('extracts status and server message from response-like error', () => {
    const normalized = normalizeApiError({
      response: {
        status: 400,
        data: {
          message: 'Bad request',
        },
      },
    })

    expect(normalized.status).toBe(400)
    expect(normalized.message).toBe('Bad request')
  })

  it('falls back to generic message for unknown errors', () => {
    const normalized = normalizeApiError(new Error('boom'))

    expect(normalized.status).toBeUndefined()
    expect(normalized.message).toBe('boom')
  })
})
