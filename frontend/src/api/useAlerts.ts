import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { api } from './client'
import type { components } from './schema'

export type AlertFeed = components['schemas']['AlertFeed']
export type AlertItem = components['schemas']['AlertItem']
export type AlertStatus = components['schemas']['AlertStatusUpdate']['status']

/** useAlerts fetches the ranked attention feed for currently open alerts. */
export function useAlerts(limit = 20) {
  return useQuery({
    queryKey: ['alerts', limit],
    queryFn: async (): Promise<AlertFeed> => {
      const { data, error } = await api.GET('/api/v1/alerts', { params: { query: { limit } } })
      if (error || !data) {
        throw new Error('failed to load alerts')
      }
      return data
    },
  })
}

/** useUpdateAlertStatus resolves or dismisses an alert from the attention feed. */
export function useUpdateAlertStatus() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (vars: { id: string; status: AlertStatus }): Promise<AlertItem> => {
      const { data, error } = await api.PATCH('/api/v1/alerts/{alertId}', {
        params: { path: { alertId: vars.id } },
        body: { status: vars.status },
      })
      if (error || !data) {
        throw new Error('failed to update alert')
      }
      return data
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['alerts'] })
    },
  })
}
