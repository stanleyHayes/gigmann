import { useMemo, useState } from 'react'
import Button from '@mui/material/Button'
import Container from '@mui/material/Container'
import CssBaseline from '@mui/material/CssBaseline'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import { ThemeProvider } from '@mui/material/styles'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

import { useBrief } from './api/useBrief'
import { DailyBrief } from './components/DailyBrief'
import { buildTheme, type ThemeMode } from './theme'

const queryClient = new QueryClient()

function Home() {
  const { data, isLoading, isError } = useBrief()
  return <DailyBrief brief={data} isLoading={isLoading} isError={isError} />
}

export function App() {
  const [mode, setMode] = useState<ThemeMode>('light')
  const theme = useMemo(() => buildTheme(mode), [mode])

  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <Container maxWidth="sm" sx={{ py: 6 }}>
          <Stack spacing={3}>
            <Stack direction="row" sx={{ justifyContent: 'space-between', alignItems: 'center' }}>
              <Typography variant="h2">Gigmann Cockpit</Typography>
              <Button
                variant="outlined"
                size="small"
                onClick={() => setMode((m) => (m === 'light' ? 'dark' : 'light'))}
              >
                {mode === 'light' ? 'Dark' : 'Light'}
              </Button>
            </Stack>
            <Home />
          </Stack>
        </Container>
      </ThemeProvider>
    </QueryClientProvider>
  )
}
