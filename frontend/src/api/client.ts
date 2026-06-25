import createClient from 'openapi-fetch'

import { getToken, setToken } from '../auth/authStore'
import type { paths } from './schema'

// Typed API client generated from backend/api/openapi.yaml.
// Regenerate types with `npm run gen:api` (or `make generate` from the repo root).
const baseUrl = import.meta.env.VITE_API_BASE_URL ?? '/'

export const api = createClient<paths>({ baseUrl })

// Attach the access token to every request and force re-login on 401.
api.use({
  onRequest({ request }) {
    const token = getToken()
    if (token) {
      request.headers.set('Authorization', `Bearer ${token}`)
    }
    return request
  },
  onResponse({ response }) {
    if (response.status === 401) {
      setToken(null)
    }
    return response
  },
})
