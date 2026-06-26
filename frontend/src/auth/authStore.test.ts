import { afterEach, describe, expect, it, vi } from 'vitest'

import { clearSession, getRefreshToken, getToken, setSession, setToken, subscribeToken } from './authStore'

afterEach(() => clearSession())

describe('authStore', () => {
  it('stores and clears the access token', () => {
    setToken('abc')
    expect(getToken()).toBe('abc')
    setToken(null)
    expect(getToken()).toBeNull()
  })

  it('stores and clears a full session', () => {
    setSession('access', 'refresh')
    expect(getToken()).toBe('access')
    expect(getRefreshToken()).toBe('refresh')
    clearSession()
    expect(getToken()).toBeNull()
    expect(getRefreshToken()).toBeNull()
  })

  it('notifies and unsubscribes listeners', () => {
    const fn = vi.fn()
    const unsubscribe = subscribeToken(fn)
    setToken('x')
    expect(fn).toHaveBeenCalledTimes(1)
    unsubscribe()
    setToken('y')
    expect(fn).toHaveBeenCalledTimes(1)
  })
})
