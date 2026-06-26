import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { clearSession, getRefreshToken, getToken, setSession } from './authStore'
import { refreshSession } from './refreshSession'

describe('refreshSession', () => {
  beforeEach(() => clearSession())
  afterEach(() => {
    clearSession()
    vi.unstubAllGlobals()
  })

  it('returns false (no fetch) when there is no refresh token', async () => {
    const fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
    expect(await refreshSession()).toBe(false)
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('stores the rotated session on success', async () => {
    setSession('old-access', 'old-refresh')
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({ ok: true, json: async () => ({ token: 'new-access', refresh_token: 'new-refresh' }) }),
    )
    expect(await refreshSession()).toBe(true)
    expect(getToken()).toBe('new-access')
    expect(getRefreshToken()).toBe('new-refresh')
  })

  it('clears the session when the refresh is rejected', async () => {
    setSession('a', 'b')
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, json: async () => ({}) }))
    expect(await refreshSession()).toBe(false)
    expect(getToken()).toBeNull()
    expect(getRefreshToken()).toBeNull()
  })
})
