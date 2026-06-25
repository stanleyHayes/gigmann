import Chip from '@mui/material/Chip'

import { statusColors, monoFont } from '../theme'

export type FacilityStatus = 'good' | 'watch' | 'critical'

const labelByStatus: Record<FacilityStatus, string> = {
  good: 'GOOD',
  watch: 'WATCH',
  critical: 'CRITICAL',
}

/**
 * StatusChip shows a facility's AI-assessed health. The status is conveyed by
 * both colour and an uppercase mono label (a11y: never colour alone). When a
 * `label` is given it is prefixed (e.g. "Tafo · CRITICAL"); otherwise the chip
 * shows just the status word.
 */
export function StatusChip({ status, label }: { status: FacilityStatus; label?: string }) {
  const text = label ? `${label} · ${labelByStatus[status]}` : labelByStatus[status]
  return (
    <Chip
      label={text}
      sx={{
        fontFamily: monoFont,
        fontSize: 12,
        color: '#fff',
        bgcolor: statusColors[status],
      }}
    />
  )
}
