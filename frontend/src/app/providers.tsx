import { useMemo, useState, type ReactNode } from 'react'
import CssBaseline from '@mui/material/CssBaseline'
import GlobalStyles from '@mui/material/GlobalStyles'
import { ThemeProvider } from '@mui/material/styles'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

import { buildTheme, type ThemeMode } from '../theme'
import { ColorModeContext, type ColorModeContextValue } from './colorMode'

const queryClient = new QueryClient()

/** AppProviders wires data fetching, theming, and the light/dark colour mode. */
export function AppProviders({ children }: { children: ReactNode }) {
  const [mode, setMode] = useState<ThemeMode>('light')

  const colorMode = useMemo<ColorModeContextValue>(
    () => ({ mode, toggle: () => setMode((m) => (m === 'light' ? 'dark' : 'light')) }),
    [mode],
  )
  const theme = useMemo(() => buildTheme(mode), [mode])

  return (
    <QueryClientProvider client={queryClient}>
      <ColorModeContext.Provider value={colorMode}>
        <ThemeProvider theme={theme}>
          <CssBaseline />
          <GlobalStyles
            styles={{
              '@media (prefers-reduced-motion: reduce)': {
                '*, *::before, *::after': {
                  animationDuration: '0.01ms !important',
                  animationIterationCount: '1 !important',
                  transitionDuration: '0.01ms !important',
                  scrollBehavior: 'auto !important',
                },
              },
            }}
          />
          {children}
        </ThemeProvider>
      </ColorModeContext.Provider>
    </QueryClientProvider>
  )
}
