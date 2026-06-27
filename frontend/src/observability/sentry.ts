/**
 * initErrorTracking lazily loads Sentry only when VITE_SENTRY_DSN is configured,
 * so the SDK is a separate chunk that never ships/loads in the default build.
 */
export function initErrorTracking(): void {
  const dsn = import.meta.env.VITE_SENTRY_DSN as string | undefined
  if (!dsn) {
    return
  }
  void import('@sentry/react').then((Sentry) => {
    Sentry.init({ dsn, environment: import.meta.env.MODE, tracesSampleRate: 0 })
  })
}
