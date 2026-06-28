import { useCallback, useEffect, useState } from 'react'

import { api } from './client'

/** Converts a base64url VAPID key to the Uint8Array the Push API expects. The
 *  array is backed by a plain ArrayBuffer so it satisfies BufferSource. */
export function urlBase64ToUint8Array(base64: string): Uint8Array<ArrayBuffer> {
  const padding = '='.repeat((4 - (base64.length % 4)) % 4)
  const normalized = (base64 + padding).replace(/-/g, '+').replace(/_/g, '/')
  const raw = atob(normalized)
  const out = new Uint8Array(new ArrayBuffer(raw.length))
  for (let i = 0; i < raw.length; i += 1) out[i] = raw.charCodeAt(i)
  return out
}

export interface PushState {
  /** The browser exposes Service Worker + Push + Notification APIs. */
  supported: boolean
  /** The server has VAPID configured (push is usable). */
  available: boolean
  /** The user currently has an active subscription. */
  enabled: boolean
  busy: boolean
  error: string | null
  enable: () => Promise<void>
  disable: () => Promise<void>
}

function browserSupportsPush(): boolean {
  return (
    typeof navigator !== 'undefined' &&
    'serviceWorker' in navigator &&
    typeof window !== 'undefined' &&
    'PushManager' in window &&
    'Notification' in window
  )
}

/**
 * usePush wires the browser Web Push subscription to the backend. Critical-only
 * alerts are delivered via the service worker (see public/push-sw.js). The whole
 * feature is hidden when the browser lacks support or the server has no VAPID key.
 */
export function usePush(): PushState {
  const supported = browserSupportsPush()
  const [vapidKey, setVapidKey] = useState('')
  const [enabled, setEnabled] = useState(false)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!supported) return
    let active = true
    void (async () => {
      const { data } = await api.GET('/api/v1/push/key')
      if (!active) return
      setVapidKey(data?.public_key ?? '')
      const reg = await navigator.serviceWorker.ready
      const sub = await reg.pushManager.getSubscription()
      if (active) setEnabled(Boolean(sub))
    })()
    return () => {
      active = false
    }
  }, [supported])

  const enable = useCallback(async () => {
    if (!supported || !vapidKey) return
    setBusy(true)
    setError(null)
    try {
      if ((await Notification.requestPermission()) !== 'granted') {
        setError('Notifications were not allowed.')
        return
      }
      const reg = await navigator.serviceWorker.ready
      const sub = await reg.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(vapidKey),
      })
      const json = sub.toJSON() as { endpoint?: string; keys?: { p256dh?: string; auth?: string } }
      const { error: apiError } = await api.POST('/api/v1/push/subscribe', {
        body: {
          endpoint: json.endpoint ?? '',
          keys: { p256dh: json.keys?.p256dh ?? '', auth: json.keys?.auth ?? '' },
        },
      })
      if (apiError) {
        setError('Could not register for notifications.')
        return
      }
      setEnabled(true)
    } finally {
      setBusy(false)
    }
  }, [supported, vapidKey])

  const disable = useCallback(async () => {
    if (!supported) return
    setBusy(true)
    setError(null)
    try {
      const reg = await navigator.serviceWorker.ready
      const sub = await reg.pushManager.getSubscription()
      if (sub) {
        await api.POST('/api/v1/push/unsubscribe', { body: { endpoint: sub.endpoint } })
        await sub.unsubscribe()
      }
      setEnabled(false)
    } finally {
      setBusy(false)
    }
  }, [supported])

  return { supported, available: Boolean(vapidKey), enabled, busy, error, enable, disable }
}
