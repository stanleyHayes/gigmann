import { useMutation } from '@tanstack/react-query'

import { api } from './client'

const genericResetMessage = 'If an account exists, password reset instructions are ready.'

export function usePasswordResetRequest() {
  return useMutation({
    mutationFn: async (email: string) => {
      const { data, error } = await api.POST('/api/v1/auth/password-reset/request', { body: { email } })
      if (error) {
        throw error
      }
      return data ?? { message: genericResetMessage }
    },
  })
}

export function usePasswordResetConfirm() {
  return useMutation({
    mutationFn: async (vars: { token: string; password: string }) => {
      const { error } = await api.POST('/api/v1/auth/password-reset/confirm', { body: vars })
      if (error) {
        throw error
      }
    },
  })
}
