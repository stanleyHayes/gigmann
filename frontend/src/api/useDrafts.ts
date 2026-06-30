import { useMutation } from '@tanstack/react-query'

import { api } from './client'
import type { components } from './schema'

export type Draft = components['schemas']['Draft']
export type DraftRequest = components['schemas']['DraftRequest']

/** useCreateDraft generates an unsent executive message or summary draft. */
export function useCreateDraft() {
  return useMutation({
    mutationFn: async (body: DraftRequest): Promise<Draft> => {
      const { data, error } = await api.POST('/api/v1/drafts', { body })
      if (error || !data) {
        throw new Error('failed to create draft')
      }
      return data
    },
  })
}
