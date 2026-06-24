import createClient from 'openapi-fetch'
import type { paths } from './schema'

// Typed API client generated from backend/api/openapi.yaml.
// Regenerate types with `npm run gen:api` (or `make generate` from the repo root).
const baseUrl = import.meta.env.VITE_API_BASE_URL ?? '/'

export const api = createClient<paths>({ baseUrl })
