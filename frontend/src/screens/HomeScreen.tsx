import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import { useNavigate } from 'react-router-dom'

import { useBrief } from '../api/useBrief'
import { DailyBrief } from '../components/DailyBrief'

/** HomeScreen is the hero "Today" view — the AI-narrated Daily Brief. */
export function HomeScreen() {
  const { data, isLoading, isError } = useBrief()
  const navigate = useNavigate()

  // A suggested action on a brief item jumps to Ask with the question prefilled.
  const askAbout = (action: string, facilityId: string) => {
    navigate('/ask', { state: { question: `${action} — ${facilityId}` } })
  }

  return (
    <Stack spacing={3}>
      <Typography variant="h1" sx={{ fontSize: { xs: '2rem', md: '2.5rem' } }}>
        The Brief
      </Typography>
      <DailyBrief brief={data} isLoading={isLoading} isError={isError} onAction={askAbout} />
    </Stack>
  )
}
