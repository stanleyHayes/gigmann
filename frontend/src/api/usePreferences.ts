import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { api } from './client'
import type { components } from './schema'

export type Preferences = components['schemas']['Preferences']

/** usePreferences loads the current user's personalisation preferences. */
export function usePreferences() {
  return useQuery({
    queryKey: ['preferences'],
    queryFn: async (): Promise<Preferences> => {
      const { data, error } = await api.GET('/api/v1/me/preferences')
      if (error || !data) {
        throw new Error('failed to load preferences')
      }
      return data
    },
  })
}

/** useSavePreferences persists updated preferences and refreshes the cache. */
export function useSavePreferences() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (prefs: Preferences): Promise<Preferences> => {
      const { data, error } = await api.PATCH('/api/v1/me/preferences', { body: prefs })
      if (error || !data) {
        throw new Error('failed to save preferences')
      }
      return data
    },
    onSuccess: (data) => queryClient.setQueryData(['preferences'], data),
  })
}
