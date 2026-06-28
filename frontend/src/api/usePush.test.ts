import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { urlBase64ToUint8Array, usePush } from './usePush'

vi.mock('./client', () => ({
  api: { GET: vi.fn(), POST: vi.fn() },
}))
import { api } from './client'

const mockedGet = vi.mocked(api.GET)
const mockedPost = vi.mocked(api.POST)

afterEach(() => {
  vi.unstubAllGlobals()
  vi.clearAllMocks()
})

describe('urlBase64ToUint8Array', () => {
  it('decodes base64url to bytes', () => {
    // "aGVsbG8" is base64url for "hello".
    expect(Array.from(urlBase64ToUint8Array('aGVsbG8'))).toEqual([104, 101, 108, 108, 111])
  })
})

describe('usePush', () => {
  it('reports unsupported in a plain environment', () => {
    const { result } = renderHook(() => usePush())
    expect(result.current.supported).toBe(false)
    expect(result.current.available).toBe(false)
  })

  it('subscribes through the push manager and registers with the API', async () => {
    const subscription = {
      endpoint: 'https://push.example.com/x',
      toJSON: () => ({ endpoint: 'https://push.example.com/x', keys: { p256dh: 'p', auth: 'a' } }),
      unsubscribe: vi.fn().mockResolvedValue(true),
    }
    const subscribe = vi.fn().mockResolvedValue(subscription)
    const getSubscription = vi.fn().mockResolvedValue(null)
    const registration = { pushManager: { subscribe, getSubscription } }

    vi.stubGlobal('navigator', { serviceWorker: { ready: Promise.resolve(registration) } })
    vi.stubGlobal('PushManager', class {})
    vi.stubGlobal('Notification', { requestPermission: vi.fn().mockResolvedValue('granted') })
    mockedGet.mockResolvedValue({ data: { public_key: 'aGVsbG8' } } as never)
    mockedPost.mockResolvedValue({ error: undefined } as never)

    const { result } = renderHook(() => usePush())
    expect(result.current.supported).toBe(true)
    await waitFor(() => expect(result.current.available).toBe(true))

    await act(async () => {
      await result.current.enable()
    })

    expect(subscribe).toHaveBeenCalledWith(
      expect.objectContaining({ userVisibleOnly: true }),
    )
    expect(mockedPost).toHaveBeenCalledWith('/api/v1/push/subscribe', {
      body: { endpoint: 'https://push.example.com/x', keys: { p256dh: 'p', auth: 'a' } },
    })
    expect(result.current.enabled).toBe(true)
  })

  it('does not subscribe when permission is denied', async () => {
    const subscribe = vi.fn()
    vi.stubGlobal('navigator', {
      serviceWorker: { ready: Promise.resolve({ pushManager: { subscribe, getSubscription: vi.fn().mockResolvedValue(null) } }) },
    })
    vi.stubGlobal('PushManager', class {})
    vi.stubGlobal('Notification', { requestPermission: vi.fn().mockResolvedValue('denied') })
    mockedGet.mockResolvedValue({ data: { public_key: 'aGVsbG8' } } as never)

    const { result } = renderHook(() => usePush())
    await waitFor(() => expect(result.current.available).toBe(true))
    await act(async () => {
      await result.current.enable()
    })

    expect(subscribe).not.toHaveBeenCalled()
    expect(result.current.enabled).toBe(false)
    expect(result.current.error).toMatch(/not allowed/i)
  })
})
