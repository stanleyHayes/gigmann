import createClient from 'openapi-fetch'

import { clearSession, getToken } from '../auth/authStore'
import { refreshSession } from '../auth/refreshSession'
import type { paths } from './schema'

// Typed API client generated from backend/api/openapi.yaml.
// Regenerate types with `npm run gen:api` (or `make generate` from the repo root).
const baseUrl = import.meta.env.VITE_API_BASE_URL ?? '/'

export const api = createClient<paths>({ baseUrl })

api.use({
  onRequest({ request }) {
    const token = getToken()
    if (token) {
      request.headers.set('Authorization', `Bearer ${token}`)
    }
    return request
  },
  async onResponse({ request, response }) {
    // On a 401 for a business call, try to rotate the refresh token once and
    // replay the request transparently; if that fails, drop to the login screen.
    if (response.status === 401 && !request.url.includes('/api/v1/auth/')) {
      if (await refreshSession()) {
        request.headers.set('Authorization', `Bearer ${getToken()}`)
        return fetch(request)
      }
      clearSession()
    }
    return response
  },
})
