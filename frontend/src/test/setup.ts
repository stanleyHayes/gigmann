import '@testing-library/jest-dom'

// Node 22+ exposes an experimental global `localStorage` that is unconfigured
// (no `--localstorage-file`) and shadows jsdom's implementation, so app code that
// reads it throws `Cannot read properties of undefined`. Install a deterministic
// in-memory Storage when the ambient one is missing/broken, so unit tests behave
// identically on every Node version (CI pins Node 20, where jsdom's own works).
function memoryStorage(): Storage {
  let store: Record<string, string> = {}
  return {
    get length() {
      return Object.keys(store).length
    },
    clear() {
      store = {}
    },
    getItem(key: string) {
      return Object.prototype.hasOwnProperty.call(store, key) ? store[key] : null
    },
    key(index: number) {
      return Object.keys(store)[index] ?? null
    },
    removeItem(key: string) {
      delete store[key]
    },
    setItem(key: string, value: string) {
      store[key] = String(value)
    },
  }
}

function ensureStorage(name: 'localStorage' | 'sessionStorage') {
  try {
    const existing = (globalThis as Record<string, unknown>)[name] as Storage | undefined
    if (existing && typeof existing.getItem === 'function') {
      existing.clear()
      return
    }
  } catch {
    /* accessing the unconfigured Node global can throw — fall through to install */
  }
  Object.defineProperty(globalThis, name, {
    value: memoryStorage(),
    configurable: true,
    writable: true,
  })
}

ensureStorage('localStorage')
ensureStorage('sessionStorage')
