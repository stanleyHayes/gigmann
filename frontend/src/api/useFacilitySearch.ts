import { useQuery } from '@tanstack/react-query'

import { api } from './client'
import type { components } from './schema'

export type FacilityMatch = components['schemas']['FacilityMatch']

const MIN_QUERY = 2
const SEARCH_LIMIT = 6

/**
 * useFacilitySearch resolves a natural-language phrase to facilities via the
 * vector-search endpoint. It stays idle until the (trimmed) query is long enough,
 * and keeps the previous results visible while the next query is in flight.
 */
export function useFacilitySearch(query: string) {
  const q = query.trim()
  return useQuery({
    queryKey: ['facility-search', q],
    enabled: q.length >= MIN_QUERY,
    placeholderData: (prev) => prev,
    queryFn: async (): Promise<FacilityMatch[]> => {
      const { data, error } = await api.GET('/api/v1/facilities/search', {
        params: { query: { q, limit: SEARCH_LIMIT } },
      })
      if (error || !data) {
        throw new Error('facility search failed')
      }
      return data.matches
    },
  })
}
