import { useMutation } from '@tanstack/react-query'

import { api } from './client'
import type { components } from './schema'

export type MfaEnrollment = components['schemas']['MfaEnrollment']

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
    mutationFn: async (vars: { secret: string; code: string }): Promise<true> => {
      const { error } = await api.POST('/api/v1/auth/mfa/confirm', { body: vars })
      if (error) {
        throw new Error('invalid code')
      }
      return true
    },
  })
}
