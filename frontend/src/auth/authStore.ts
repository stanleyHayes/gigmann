const ACCESS_KEY = 'gigmann.token'
const REFRESH_KEY = 'gigmann.refresh'
const subscribers = new Set<() => void>()

function notify() {
  subscribers.forEach((fn) => fn())
}

/** getToken reads the persisted access token (null when signed out). */
export function getToken(): string | null {
  return localStorage.getItem(ACCESS_KEY)
}

/** getRefreshToken reads the persisted refresh token. */
export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_KEY)
}

/** setToken sets (or clears) just the access token and notifies subscribers. */
export function setToken(token: string | null): void {
  if (token) {
    localStorage.setItem(ACCESS_KEY, token)
  } else {
    localStorage.removeItem(ACCESS_KEY)
  }
  notify()
}

/** setSession persists both tokens (login or rotation) and notifies subscribers. */
export function setSession(accessToken: string, refreshToken: string): void {
  localStorage.setItem(ACCESS_KEY, accessToken)
  localStorage.setItem(REFRESH_KEY, refreshToken)
  notify()
}

/** clearSession removes both tokens (logout / failed refresh). */
export function clearSession(): void {
  localStorage.removeItem(ACCESS_KEY)
  localStorage.removeItem(REFRESH_KEY)
  notify()
}

/** subscribeToken registers a listener for token changes. */
export function subscribeToken(fn: () => void): () => void {
  subscribers.add(fn)
  return () => {
    subscribers.delete(fn)
  }
}
