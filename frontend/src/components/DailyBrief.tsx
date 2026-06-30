import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Paper from '@mui/material/Paper'
import Skeleton from '@mui/material/Skeleton'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import AddTaskOutlined from '@mui/icons-material/AddTaskOutlined'
import TipsAndUpdatesOutlined from '@mui/icons-material/TipsAndUpdatesOutlined'

import { motion, useReducedMotion } from 'framer-motion'

import type { Brief } from '../api/useBrief'
import { StatusChip, type FacilityStatus } from './StatusChip'
import { fmt } from '../i18n/locale'
import { t } from '../i18n/messages'

type BriefItem = Brief['items'][number]

type Props = {
  brief?: Brief
  isLoading: boolean
  isError: boolean
  onAction?: (action: string, facilityId: string) => void
  onTask?: (item: BriefItem) => void
}

/** DailyBrief is the hero surface: the morning brief, worst item first. */
export function DailyBrief({ brief, isLoading, isError, onAction, onTask }: Props) {
  const reduceMotion = useReducedMotion()
  if (isLoading) {
    return (
      <Box data-testid="brief-skeleton">
        <Stack spacing={2}>
          <Skeleton variant="rounded" height={120} />
          <Skeleton variant="rounded" height={132} />
          <Skeleton variant="rounded" height={132} />
        </Stack>
      </Box>
    )
  }

  if (isError || !brief) {
    return <Alert severity="error">Couldn&apos;t load the brief. Try again shortly.</Alert>
  }

  return (
    <Stack spacing={2.25}>
      <Paper
        variant="outlined"
        sx={{
          p: { xs: 2, md: 2.5 },
          bgcolor: 'background.paper',
          borderLeft: 4,
          borderLeftColor: 'primary.main',
        }}
      >
        <Stack direction="row" spacing={1.5} sx={{ alignItems: 'flex-start' }}>
          <TipsAndUpdatesOutlined color="primary" aria-hidden="true" />
          <Box>
            <Typography variant="overline" sx={{ color: 'text.secondary', fontWeight: 800, letterSpacing: 0 }}>
              Chief-of-staff readout
            </Typography>
            <Typography variant="body1" sx={{ mt: 0.5, lineHeight: 1.8 }}>
              {brief.prose}
            </Typography>
          </Box>
        </Stack>
      </Paper>
      {brief.items.map((item, i) => (
        <motion.div
          key={`${item.facility_id}-${i}`}
          initial={reduceMotion ? false : { opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.2, delay: reduceMotion ? 0 : i * 0.06 }}
        >
          <Paper
            variant="outlined"
            sx={{
              position: 'relative',
              overflow: 'hidden',
              p: { xs: 2, md: 2.5 },
              transition: 'transform 160ms ease, border-color 160ms ease',
              '&:hover': {
                transform: 'translateY(-2px)',
                borderColor: 'primary.light',
              },
            }}
          >
            <Box
              aria-hidden="true"
              sx={{
                position: 'absolute',
                left: 0,
                top: 18,
                bottom: 18,
                width: 4,
                borderRadius: '0 8px 8px 0',
                bgcolor: (theme) =>
                  item.severity === 'critical'
                    ? 'error.main'
                    : item.severity === 'watch'
                      ? 'warning.main'
                      : theme.palette.primary.main,
              }}
            />
            <Stack spacing={1.25} sx={{ pl: 0.5 }}>
              <Stack direction="row" spacing={1} sx={{ justifyContent: 'space-between', alignItems: 'flex-start', gap: 1 }}>
                <StatusChip status={item.severity as FacilityStatus} label={item.facility_id} />
              </Stack>
              <Typography variant="h6">{item.headline}</Typography>
              {item.explanation ? (
                <Typography variant="body2" color="text.secondary" sx={{ lineHeight: 1.7 }}>
                  {item.explanation}
                </Typography>
              ) : null}
              <Stack direction="row" spacing={1} sx={{ flexWrap: 'wrap', gap: 1, pt: 0.5 }}>
                {item.suggested_actions?.map((action, ai) => (
                  <Button
                    key={`${action}-${ai}`}
                    size="small"
                    variant="outlined"
                    aria-label={`${action} for ${item.facility_id}`}
                    onClick={onAction ? () => onAction(action, item.facility_id) : undefined}
                  >
                    {action}
                  </Button>
                ))}
                {onTask ? (
                  <Button
                    size="small"
                    variant="text"
                    startIcon={<AddTaskOutlined fontSize="small" />}
                    aria-label={`Turn ${item.facility_id} into a task`}
                    onClick={() => onTask(item)}
                  >
                    Turn into task
                  </Button>
                ) : null}
              </Stack>
            </Stack>
          </Paper>
        </motion.div>
      ))}
      <Typography variant="caption" color="text.secondary" data-testid="brief-source">
        {brief.model.toLowerCase().includes('claude') ? t('brief.source.claude') : t('brief.source.local')}
        {brief.generated_at ? ` · ${fmt.dateTime(brief.generated_at)}` : ''}
      </Typography>
    </Stack>
  )
}
