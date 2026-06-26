import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import { useRouteError } from 'react-router-dom'

/** RouteError is the router-level error boundary: it catches render errors and
 *  failed lazy-chunk loads (e.g. after a redeploy) and offers a reload. */
export function RouteError() {
  const error = useRouteError()
  const message = error instanceof Error ? error.message : 'An unexpected error occurred.'
  return (
    <Box sx={{ minHeight: '60vh', display: 'flex', alignItems: 'center', justifyContent: 'center', p: 2 }}>
      <Stack spacing={2} sx={{ maxWidth: 420 }}>
        <Typography variant="h2" sx={{ fontSize: '1.75rem' }}>
          Something went wrong
        </Typography>
        <Alert severity="error">{message}</Alert>
        <Button variant="contained" onClick={() => window.location.reload()}>
          Reload
        </Button>
      </Stack>
    </Box>
  )
}
