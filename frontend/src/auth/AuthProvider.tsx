import { useState, useSyncExternalStore, type ReactNode } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { api } from '../api/client'
import { AuthContext, type AuthUser, type AuthValue } from './authContext'
import { getToken, setToken, subscribeToken } from './authStore'

/** AuthProvider holds the session: the token, the current user, and login/logout. */
export function AuthProvider({ children }: { children: ReactNode }) {
  const token = useSyncExternalStore(subscribeToken, getToken, getToken)
  const queryClient = useQueryClient()
  const [loginError, setLoginError] = useState<string | null>(null)

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
    mutationFn: async (vars: { email: string; password: string }): Promise<string> => {
      const { data, error } = await api.POST('/api/v1/auth/login', { body: vars })
      if (error || !data) {
        throw new Error('invalid_credentials')
      }
      return data.token
    },
    onSuccess: (newToken) => {
      setLoginError(null)
      setToken(newToken)
    },
    onError: () => setLoginError('Invalid email or password.'),
  })

  const value: AuthValue = {
    user: meQuery.data,
    isAuthenticated: !!token,
    login: (email, password) => loginMutation.mutate({ email, password }),
    logout: () => {
      setToken(null)
      queryClient.clear()
    },
    loginPending: loginMutation.isPending,
    loginError,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
