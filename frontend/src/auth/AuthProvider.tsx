import { useState, useSyncExternalStore, type ReactNode } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { api } from '../api/client'
import { AuthContext, type AuthUser, type AuthValue } from './authContext'
import { clearSession, getRefreshToken, getToken, setSession, subscribeToken } from './authStore'

/** AuthProvider holds the session: the token, the current user, and login/logout. */
export function AuthProvider({ children }: { children: ReactNode }) {
  const token = useSyncExternalStore(subscribeToken, getToken, getToken)
  const queryClient = useQueryClient()
  const [loginError, setLoginError] = useState<string | null>(null)
  const [mfaRequired, setMfaRequired] = useState(false)

  const meQuery = useQuery({
    queryKey: ['auth', 'me', token],
    enabled: !!token,
    retry: false,
    queryFn: async (): Promise<AuthUser> => {
      const { data, error } = await api.GET('/api/v1/auth/me')
      if (error || !data) {
        throw new Error('unauthenticated')
      }
      return data
    },
  })

  const loginMutation = useMutation({
    mutationFn: async (vars: { email: string; password: string; code?: string }) => {
      const { data, error } = await api.POST('/api/v1/auth/login', { body: vars })
      if (error) {
        throw error
      }
      if (!data) {
        throw new Error('invalid_credentials')
      }
      return data
    },
    onSuccess: (session) => {
      setLoginError(null)
      setMfaRequired(false)
      setSession(session.token, session.refresh_token)
    },
    onError: (err: unknown) => {
      if ((err as { error?: string } | null)?.error === 'mfa_required') {
        setMfaRequired(true)
        setLoginError(null)
      } else {
        setMfaRequired(false)
        setLoginError('Invalid email or password.')
      }
    },
  })

  const value: AuthValue = {
    user: meQuery.data,
    isAuthenticated: !!token,
    mfaRequired,
    login: (email, password, code) => loginMutation.mutate({ email, password, code: code || undefined }),
    logout: () => {
      const refreshToken = getRefreshToken()
      if (refreshToken) {
        void api.POST('/api/v1/auth/logout', { body: { refresh_token: refreshToken } })
      }
      clearSession()
      queryClient.clear()
    },
    loginPending: loginMutation.isPending,
    loginError,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
