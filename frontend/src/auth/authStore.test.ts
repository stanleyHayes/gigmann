import { afterEach, describe, expect, it, vi } from 'vitest'

import { getToken, setToken, subscribeToken } from './authStore'

afterEach(() => setToken(null))

describe('authStore', () => {
  it('stores and clears the token', () => {
    setToken('abc')
    expect(getToken()).toBe('abc')
    setToken(null)
    expect(getToken()).toBeNull()
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
