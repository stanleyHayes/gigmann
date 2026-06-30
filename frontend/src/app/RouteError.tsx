import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import ErrorOutlineOutlined from '@mui/icons-material/ErrorOutlineOutlined'
import { useRouteError } from 'react-router-dom'

import { EmptyState } from '../components/EmptyState'

/** RouteError is the router-level error boundary: it catches render errors and
 *  failed lazy-chunk loads (e.g. after a redeploy) and offers a reload. */
export function RouteError() {
  const error = useRouteError()
  const message = error instanceof Error ? error.message : 'An unexpected error occurred.'
  return (
    <Box sx={{ minHeight: '60vh', display: 'flex', alignItems: 'center', justifyContent: 'center', p: 2 }}>
      <Box sx={{ width: '100%', maxWidth: 560 }}>
        <EmptyState
          icon={ErrorOutlineOutlined}
          title="Something went wrong"
          description={message}
          actions={
            <Button variant="contained" onClick={() => window.location.reload()}>
              Reload
            </Button>
          }
        />
      </Box>
    </Box>
  )
}
