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
 * both colour and an uppercase mono label (a11y: never colour alone).
 */
export function StatusChip({ status, label }: { status: FacilityStatus; label: string }) {
  return (
    <Chip
      label={`${label} · ${labelByStatus[status]}`}
      sx={{
        fontFamily: monoFont,
        fontSize: 12,
        color: '#fff',
        bgcolor: statusColors[status],
      }}
    />
  )
}
