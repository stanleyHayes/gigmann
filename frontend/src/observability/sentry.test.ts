import { afterEach, describe, expect, it, vi } from 'vitest'

import { initErrorTracking } from './sentry'

describe('initErrorTracking', () => {
  afterEach(() => vi.unstubAllEnvs())

  it('is a no-op when no DSN is configured', () => {
    vi.stubEnv('VITE_SENTRY_DSN', '')
    expect(() => initErrorTracking()).not.toThrow()
  })

  it('does not throw when a DSN is set', () => {
    vi.stubEnv('VITE_SENTRY_DSN', 'https://examplePublicKey@o0.ingest.sentry.io/0')
    expect(() => initErrorTracking()).not.toThrow()
  })
})
