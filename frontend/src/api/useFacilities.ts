import { useQuery } from '@tanstack/react-query'

import { api } from './client'
import type { components } from './schema'

export type Facility = components['schemas']['Facility']

/** useFacilities fetches the full facility network from the API. */
export function useFacilities() {
  return useQuery({
    queryKey: ['facilities'],
    queryFn: async (): Promise<Facility[]> => {
      const { data, error } = await api.GET('/api/v1/facilities')
      if (error || !data) {
        throw new Error('failed to load facilities')
      }
      return data.facilities
    },
  })
}
