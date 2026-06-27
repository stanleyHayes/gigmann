import { useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'

import { getToken } from '../auth/authStore'

/**
 * useLiveUpdates opens the realtime channel (WebSocket) and invalidates the
 * relevant query caches on server events (e.g. a refreshed brief). It is
 * best-effort: it silently no-ops when WebSocket is unavailable (SSR/jsdom) or
 * the user isn't authenticated, so it never breaks the app.
 */
export function useLiveUpdates(): void {
  const queryClient = useQueryClient()
  useEffect(() => {
    if (typeof WebSocket === 'undefined') {
      return
    }
    const token = getToken()
    if (!token) {
      return
    }
    const base = (import.meta.env.VITE_API_BASE_URL as string | undefined) || window.location.origin
    let url: URL
    try {
      url = new URL('/api/v1/ws', base)
    } catch {
      return
    }
    url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
    url.searchParams.set('token', token)

    let ws: WebSocket | undefined
    try {
      ws = new WebSocket(url.toString())
      ws.onmessage = (e: MessageEvent) => {
        if (e.data === 'brief.refreshed') {
          void queryClient.invalidateQueries({ queryKey: ['brief'] })
        }
      }
    } catch {
      // realtime is optional — ignore connection failures
    }
    return () => ws?.close()
  }, [queryClient])
}
