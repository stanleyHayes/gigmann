import createClient from 'openapi-fetch'

import { clearSession, getToken } from '../auth/authStore'
import { refreshSession } from '../auth/refreshSession'
import type { paths } from './schema'

// Typed API client generated from backend/api/openapi.yaml.
// Regenerate types with `npm run gen:api` (or `make generate` from the repo root).
const baseUrl = import.meta.env.VITE_API_BASE_URL ?? '/'

export const api = createClient<paths>({ baseUrl })

// A Request body can only be read once, so the 401-retry below must replay a
// clone captured before the original was sent — otherwise a mutation (POST/PATCH)
// would be retried with an empty body. The WeakMap is self-cleaning (keyed by the
// request, which is GC'd once the call settles).
const retryClones = new WeakMap<Request, Request>()

api.use({
  onRequest({ request }) {
    const token = getToken()
    if (token) {
      request.headers.set('Authorization', `Bearer ${token}`)
    }
    retryClones.set(request, request.clone())
    return request
  },
  async onResponse({ request, response }) {
    // On a 401 for a business call, try to rotate the refresh token once and
    // replay the request transparently; if that fails, drop to the login screen.
    if (response.status === 401 && !request.url.includes('/api/v1/auth/')) {
      if (await refreshSession()) {
        const retry = retryClones.get(request) ?? request
        retry.headers.set('Authorization', `Bearer ${getToken()}`)
        return fetch(retry)
      }
      clearSession()
    }
    return response
  },
})
