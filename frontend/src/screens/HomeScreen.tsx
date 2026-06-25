import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'

import { useBrief } from '../api/useBrief'
import { DailyBrief } from '../components/DailyBrief'

/** HomeScreen is the hero "Today" view — the AI-narrated Daily Brief. */
export function HomeScreen() {
  const { data, isLoading, isError } = useBrief()
  return (
    <Stack spacing={3}>
      <Typography variant="h1" sx={{ fontSize: { xs: '2rem', md: '2.5rem' } }}>
        The Brief
      </Typography>
      <DailyBrief brief={data} isLoading={isLoading} isError={isError} />
    </Stack>
  )
}
