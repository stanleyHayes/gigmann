import { useQuery } from '@tanstack/react-query'

import { api } from './client'
import type { components } from './schema'

export type NetworkMetrics = components['schemas']['NetworkMetrics']
export type Kpi = components['schemas']['Kpi']
export type MetricPoint = components['schemas']['MetricPoint']

/** useMetrics fetches the deterministic network KPIs and trends from the API. */
export function useMetrics() {
  return useQuery({
    queryKey: ['metrics'],
    queryFn: async (): Promise<NetworkMetrics> => {
      const { data, error } = await api.GET('/api/v1/metrics')
      if (error || !data) {
        throw new Error('failed to load metrics')
      }
      return data
    },
  })
}
