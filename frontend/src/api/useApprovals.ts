import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { api } from './client'
import type { components } from './schema'

export type Approval = components['schemas']['Approval']
export type Decision = 'approve' | 'decline'

/** useApprovals fetches the approvals routed to the executive. */
export function useApprovals() {
  return useQuery({
    queryKey: ['approvals'],
    queryFn: async (): Promise<Approval[]> => {
      const { data, error } = await api.GET('/api/v1/approvals')
      if (error || !data) {
        throw new Error('failed to load approvals')
      }
      return data.approvals
    },
  })
}

/** useDecideApproval records an explicit approve/decline decision. */
export function useDecideApproval() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (vars: { id: string; decision: Decision; note?: string }): Promise<Approval> => {
      const { data, error } = await api.POST('/api/v1/approvals/{approvalId}/decision', {
        params: { path: { approvalId: vars.id } },
        body: { decision: vars.decision, note: vars.note },
      })
      if (error || !data) {
        throw new Error('decision failed')
      }
      return data
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['approvals'] })
    },
  })
}
