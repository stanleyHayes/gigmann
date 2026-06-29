import { useMutation, useQueryClient } from '@tanstack/react-query'

import { api } from './client'
import type { components } from './schema'

export type MfaEnrollment = components['schemas']['MfaEnrollment']
export type MfaRecoveryCodes = components['schemas']['MfaRecoveryCodes']

/** useMfaEnroll begins TOTP enrollment (returns a secret + otpauth URI). */
export function useMfaEnroll() {
  return useMutation({
    mutationFn: async (): Promise<MfaEnrollment> => {
      const { data, error } = await api.POST('/api/v1/auth/mfa/enroll', {})
      if (error || !data) {
        throw new Error('failed to start enrollment')
      }
      return data
    },
  })
}

/** useMfaConfirm confirms enrollment with a code, activating MFA. */
export function useMfaConfirm() {
  return useMutation({
    mutationFn: async (vars: { secret: string; code: string }): Promise<MfaRecoveryCodes> => {
      const { data, error } = await api.POST('/api/v1/auth/mfa/confirm', { body: vars })
      if (error || !data) {
        throw new Error('invalid code')
      }
      return data
    },
  })
}

/** useMfaDisable disables MFA after a current authenticator or recovery code. */
export function useMfaDisable() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (vars: { code: string }): Promise<true> => {
      const { error } = await api.POST('/api/v1/auth/mfa/disable', { body: vars })
      if (error) {
        throw new Error('invalid code')
      }
      return true
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['auth', 'me'] })
    },
  })
}
