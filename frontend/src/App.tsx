import { useMemo, useState } from 'react'
import CssBaseline from '@mui/material/CssBaseline'
import { ThemeProvider } from '@mui/material/styles'
import Container from '@mui/material/Container'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import Button from '@mui/material/Button'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { buildTheme, type ThemeMode } from './theme'
import { StatusChip } from './components/StatusChip'
import { ButtonLoadingDots } from './components/ButtonLoadingDots'

const queryClient = new QueryClient()

export function App() {
  const [mode, setMode] = useState<ThemeMode>('light')
  const theme = useMemo(() => buildTheme(mode), [mode])

  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <Container maxWidth="sm" sx={{ py: 6 }}>
          <Stack spacing={3}>
            <Typography variant="h2">Gigmann Executive Cockpit</Typography>
            <Typography color="text.secondary">
              Scaffold ready. The Daily Brief and the living network arrive in epics E6/E7.
            </Typography>
            <Stack direction="row" spacing={1} sx={{ flexWrap: 'wrap' }}>
              <StatusChip status="critical" label="Tafo Maternity" />
              <StatusChip status="watch" label="Asokwa" />
              <StatusChip status="good" label="Adansi" />
            </Stack>
            <Button variant="contained" disabled>
              <ButtonLoadingDots /> Generating brief
            </Button>
            <Button variant="outlined" onClick={() => setMode((m) => (m === 'light' ? 'dark' : 'light'))}>
              Toggle {mode === 'light' ? 'dark' : 'light'} mode
            </Button>
          </Stack>
        </Container>
      </ThemeProvider>
    </QueryClientProvider>
  )
}
