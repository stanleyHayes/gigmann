import { clearSession, getRefreshToken, setSession } from './authStore'

const baseUrl = (import.meta.env.VITE_API_BASE_URL ?? '/').replace(/\/$/, '')
const refreshUrl = `${baseUrl}/api/v1/auth/refresh`

let inFlight: Promise<boolean> | null = null

/**
 * refreshSession rotates the stored refresh token into a new access+refresh
 * pair. It uses a raw fetch (bypassing the API-client middleware, so it never
 * recurses) and de-dupes concurrent callers behind a single in-flight request.
 */
export function refreshSession(): Promise<boolean> {
  inFlight ??= doRefresh().finally(() => {
    inFlight = null
  })
  return inFlight
}

async function doRefresh(): Promise<boolean> {
  const refreshToken = getRefreshToken()
  if (!refreshToken) {
    return false
  }
  try {
    const res = await fetch(refreshUrl, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })
    if (!res.ok) {
      clearSession()
      return false
    }
    const data = (await res.json()) as { token: string; refresh_token: string }
    setSession(data.token, data.refresh_token)
    return true
  } catch {
    clearSession()
    return false
  }
}
