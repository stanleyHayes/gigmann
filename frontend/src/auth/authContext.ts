import { createContext, useContext } from 'react'

import type { components } from '../api/schema'

export type AuthUser = components['schemas']['AuthUser']

export type AuthValue = {
  user: AuthUser | undefined
  isAuthenticated: boolean
  login: (email: string, password: string) => void
  logout: () => void
  loginPending: boolean
  loginError: string | null
}

export const AuthContext = createContext<AuthValue | undefined>(undefined)

/** useAuth reads the auth context; must be used within AuthProvider. */
export function useAuth(): AuthValue {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return ctx
}
