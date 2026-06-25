import { createContext, useContext } from 'react'

import type { ThemeMode } from '../theme'

export type ColorModeContextValue = {
  mode: ThemeMode
  toggle: () => void
}

export const ColorModeContext = createContext<ColorModeContextValue | undefined>(undefined)

/** useColorMode reads the current theme mode and a toggle; must be used within AppProviders. */
export function useColorMode(): ColorModeContextValue {
  const ctx = useContext(ColorModeContext)
  if (!ctx) {
    throw new Error('useColorMode must be used within AppProviders')
  }
  return ctx
}
