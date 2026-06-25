const KEY = 'gigmann.token'
const subscribers = new Set<() => void>()

/** getToken reads the persisted access token (null when signed out). */
export function getToken(): string | null {
  return localStorage.getItem(KEY)
}

/** setToken persists (or clears) the token and notifies subscribers. */
export function setToken(token: string | null): void {
  if (token) {
    localStorage.setItem(KEY, token)
  } else {
    localStorage.removeItem(KEY)
  }
  subscribers.forEach((fn) => fn())
}

/** subscribeToken registers a listener for token changes. */
export function subscribeToken(fn: () => void): () => void {
  subscribers.add(fn)
  return () => {
    subscribers.delete(fn)
  }
}
