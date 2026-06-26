import { useMutation } from '@tanstack/react-query'

import { api } from './client'
import type { components } from './schema'

export type Answer = components['schemas']['Answer']

/** useAsk asks a natural-language question about the network (grounded answer). */
export function useAsk() {
  return useMutation({
    mutationFn: async (question: string): Promise<Answer> => {
      const { data, error } = await api.POST('/api/v1/ask', { body: { question } })
      if (error || !data) {
        throw new Error('failed to get an answer')
      }
      return data
    },
  })
}
