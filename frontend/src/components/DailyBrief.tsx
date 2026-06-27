import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Paper from '@mui/material/Paper'
import Skeleton from '@mui/material/Skeleton'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'

import { motion, useReducedMotion } from 'framer-motion'

import type { Brief } from '../api/useBrief'
import { StatusChip, type FacilityStatus } from './StatusChip'

type Props = {
  brief?: Brief
  isLoading: boolean
  isError: boolean
  onAction?: (action: string, facilityId: string) => void
}

/** DailyBrief is the hero surface: the morning brief, worst item first. */
export function DailyBrief({ brief, isLoading, isError, onAction }: Props) {
  const reduceMotion = useReducedMotion()
  if (isLoading) {
    return (
      <Box data-testid="brief-skeleton">
        <Stack spacing={2}>
          <Skeleton variant="text" width="80%" height={32} />
          <Skeleton variant="rounded" height={88} />
          <Skeleton variant="rounded" height={88} />
        </Stack>
      </Box>
    )
  }

  if (isError || !brief) {
    return <Alert severity="error">Couldn&apos;t load the brief. Try again shortly.</Alert>
  }

  return (
    <Stack spacing={2}>
      <Typography variant="body1">{brief.prose}</Typography>
      {brief.items.map((item, i) => (
        <motion.div
          key={`${item.facility_id}-${i}`}
          initial={reduceMotion ? false : { opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.2, delay: reduceMotion ? 0 : i * 0.06 }}
        >
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Stack spacing={1}>
            <StatusChip status={item.severity as FacilityStatus} label={item.facility_id} />
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              {item.headline}
            </Typography>
            {item.explanation ? (
              <Typography variant="body2" color="text.secondary">
                {item.explanation}
              </Typography>
            ) : null}
            {item.suggested_actions && item.suggested_actions.length > 0 ? (
              <Stack direction="row" spacing={1} sx={{ flexWrap: 'wrap' }}>
                {item.suggested_actions.map((action) => (
                  <Button
                    key={action}
                    size="small"
                    variant="outlined"
                    aria-label={`${action} for ${item.facility_id}`}
                    onClick={onAction ? () => onAction(action, item.facility_id) : undefined}
                  >
                    {action}
                  </Button>
                ))}
              </Stack>
            ) : null}
          </Stack>
        </Paper>
        </motion.div>
      ))}
      <Typography variant="caption" color="text.secondary" data-testid="brief-source">
        {brief.model.toLowerCase().includes('claude') ? 'Narrated by Claude' : 'Deterministic summary — AI narration unavailable'}
        {brief.generated_at ? ` · ${new Date(brief.generated_at).toLocaleString('en-GB')}` : ''}
      </Typography>
    </Stack>
  )
}
