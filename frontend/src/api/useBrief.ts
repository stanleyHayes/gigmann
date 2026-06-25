import { useQuery } from '@tanstack/react-query'

import { api } from './client'
import type { components } from './schema'

export type Brief = components['schemas']['Brief']
export type BriefItem = components['schemas']['BriefItem']

/** useBrief fetches the AI-narrated Daily Brief from the API. */
export function useBrief() {
  return useQuery({
    queryKey: ['brief'],
    queryFn: async (): Promise<Brief> => {
      const { data, error } = await api.GET('/api/v1/brief')
      if (error || !data) {
        throw new Error('failed to load the brief')
      }
      return data
    },
  })
}
