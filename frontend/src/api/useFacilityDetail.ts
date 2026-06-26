import { useQuery } from '@tanstack/react-query'

import { api } from './client'
import type { components } from './schema'

export type FacilityDetail = components['schemas']['FacilityDetail']

/** useFacilityDetail fetches one facility's drill-down (inventory, staff, alerts). */
export function useFacilityDetail(facilityId: string) {
  return useQuery({
    queryKey: ['facility', facilityId],
    enabled: facilityId !== '',
    queryFn: async (): Promise<FacilityDetail> => {
      const { data, error } = await api.GET('/api/v1/facilities/{facilityId}', {
        params: { path: { facilityId } },
      })
      if (error || !data) {
        throw new Error('failed to load facility')
      }
      return data
    },
  })
}
