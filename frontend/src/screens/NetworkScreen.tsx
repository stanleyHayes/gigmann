import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import Card from '@mui/material/Card'
import CardActionArea from '@mui/material/CardActionArea'
import CardContent from '@mui/material/CardContent'
import Skeleton from '@mui/material/Skeleton'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'

import { Link as RouterLink } from 'react-router-dom'

import { useFacilities, type Facility } from '../api/useFacilities'
import { StatusChip, type FacilityStatus } from '../components/StatusChip'
import { statusColors } from '../theme'

const RANK: Record<FacilityStatus, number> = { critical: 0, watch: 1, good: 2 }
const ORDER: FacilityStatus[] = ['critical', 'watch', 'good']
const PHRASE: Record<FacilityStatus, string> = { critical: 'critical', watch: 'to watch', good: 'healthy' }

const GRID = {
  display: 'grid',
  gap: 2,
  gridTemplateColumns: { xs: '1fr', sm: 'repeat(2, 1fr)', md: 'repeat(3, 1fr)' },
} as const

function countByStatus(facilities: Facility[]): Record<FacilityStatus, number> {
  const counts: Record<FacilityStatus, number> = { critical: 0, watch: 0, good: 0 }
  for (const f of facilities) {
    counts[f.status as FacilityStatus] += 1
  }
  return counts
}

function NetworkSummary({ facilities }: { facilities: Facility[] }) {
  const counts = countByStatus(facilities)
  const total = facilities.length
  const phrase =
    total === 0 ? 'no facilities yet' : ORDER.filter((s) => counts[s] > 0).map((s) => `${counts[s]} ${PHRASE[s]}`).join(' · ')

  return (
    <Stack spacing={1}>
      <Typography variant="body1" sx={{ fontWeight: 600 }}>
        {total} {total === 1 ? 'facility' : 'facilities'} · {phrase}
      </Typography>
      <Box
        role="img"
        aria-label={`Network health: ${phrase}`}
        sx={{ display: 'flex', height: 12, borderRadius: 1, overflow: 'hidden', bgcolor: 'action.hover' }}
      >
        {ORDER.map((s) =>
          counts[s] > 0 ? (
            <Box key={s} sx={{ flexGrow: counts[s], flexBasis: 0, bgcolor: statusColors[s] }} />
          ) : null,
        )}
      </Box>
    </Stack>
  )
}

function FacilityCard({ facility }: { facility: Facility }) {
  return (
    <Card variant="outlined" component="article" aria-label={facility.name}>
      <CardActionArea component={RouterLink} to={`/facilities/${facility.id}`}>
      <CardContent>
        <Stack spacing={1}>
          <Stack direction="row" spacing={1} sx={{ justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              {facility.name}
            </Typography>
            <StatusChip status={facility.status as FacilityStatus} />
          </Stack>
          <Typography variant="body2" color="text.secondary">
            {facility.town}, {facility.region}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {facility.beds} beds
          </Typography>
        </Stack>
      </CardContent>
      </CardActionArea>
    </Card>
  )
}

/** NetworkOverview renders the network summary and worst-first facility grid. */
export function NetworkOverview({ facilities }: { facilities: Facility[] }) {
  const sorted = [...facilities].sort(
    (a, b) => RANK[a.status as FacilityStatus] - RANK[b.status as FacilityStatus] || a.name.localeCompare(b.name),
  )
  return (
    <Stack spacing={3}>
      <NetworkSummary facilities={facilities} />
      <Box sx={GRID}>
        {sorted.map((f) => (
          <FacilityCard key={f.id} facility={f} />
        ))}
      </Box>
    </Stack>
  )
}

/** NetworkScreen is the single-pane view of the whole facility network. */
export function NetworkScreen() {
  const { data, isLoading, isError } = useFacilities()

  return (
    <Stack spacing={3}>
      <Typography variant="h1" sx={{ fontSize: { xs: '2rem', md: '2.5rem' } }}>
        Network
      </Typography>
      {isLoading ? (
        <Stack spacing={2} data-testid="network-skeleton">
          <Skeleton variant="text" width="60%" />
          <Skeleton variant="rounded" height={12} />
          <Box sx={GRID}>
            {[0, 1, 2, 3, 4, 5].map((i) => (
              <Skeleton key={i} variant="rounded" height={132} />
            ))}
          </Box>
        </Stack>
      ) : isError || !data ? (
        <Alert severity="error">Couldn&apos;t load the network. Try again shortly.</Alert>
      ) : (
        <NetworkOverview facilities={data} />
      )}
    </Stack>
  )
}
