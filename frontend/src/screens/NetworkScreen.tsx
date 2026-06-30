import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import Card from '@mui/material/Card'
import CardActionArea from '@mui/material/CardActionArea'
import CardContent from '@mui/material/CardContent'
import FormControl from '@mui/material/FormControl'
import InputLabel from '@mui/material/InputLabel'
import MenuItem from '@mui/material/MenuItem'
import Skeleton from '@mui/material/Skeleton'
import Stack from '@mui/material/Stack'
import Select from '@mui/material/Select'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'
import HubOutlined from '@mui/icons-material/HubOutlined'
import MonitorHeartOutlined from '@mui/icons-material/MonitorHeartOutlined'

import { Link as RouterLink } from 'react-router-dom'
import { useMemo, useState } from 'react'

import { useFacilities, type Facility } from '../api/useFacilities'
import { fmt } from '../i18n/locale'
import { PageHeader } from '../components/PageHeader'
import { PaginationControls, usePagination } from '../components/PaginationControls'
import { StatusChip, type FacilityStatus } from '../components/StatusChip'
import { SurfaceCard } from '../components/SurfaceCard'
import { statusColors } from '../theme'

const RANK: Record<FacilityStatus, number> = { critical: 0, watch: 1, good: 2 }
const ORDER: FacilityStatus[] = ['critical', 'watch', 'good']
const PHRASE: Record<FacilityStatus, string> = { critical: 'critical', watch: 'to watch', good: 'healthy' }
type StatusFilter = FacilityStatus | 'all'
type SortKey = 'attention' | 'name' | 'region' | 'beds' | 'revenue' | 'occupancy' | 'patients'

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
    <SurfaceCard
      title={`${total} ${total === 1 ? 'facility' : 'facilities'}`}
      description={phrase}
      icon={MonitorHeartOutlined}
    >
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
    </SurfaceCard>
  )
}

function formatRevenue(pesewas: number | undefined) {
  return pesewas == null ? 'No reading' : fmt.cedis(pesewas / 100)
}

function formatOccupancy(rate: number | undefined) {
  return rate == null ? 'No reading' : `${(rate * 100).toFixed(1)}%`
}

function sortFacilities(facilities: Facility[], sort: SortKey): Facility[] {
  return [...facilities].sort((a, b) => {
    switch (sort) {
      case 'name':
        return a.name.localeCompare(b.name)
      case 'region':
        return a.region.localeCompare(b.region) || a.name.localeCompare(b.name)
      case 'beds':
        return b.beds - a.beds || a.name.localeCompare(b.name)
      case 'revenue':
        return (b.latest_revenue_pesewas ?? -1) - (a.latest_revenue_pesewas ?? -1) || a.name.localeCompare(b.name)
      case 'occupancy':
        return (b.occupancy_rate ?? -1) - (a.occupancy_rate ?? -1) || a.name.localeCompare(b.name)
      case 'patients':
        return (b.patients_seen ?? -1) - (a.patients_seen ?? -1) || a.name.localeCompare(b.name)
      default:
        return RANK[a.status as FacilityStatus] - RANK[b.status as FacilityStatus] || a.name.localeCompare(b.name)
    }
  })
}

function NetworkControls({
  search,
  status,
  region,
  sort,
  regions,
  onSearch,
  onStatus,
  onRegion,
  onSort,
}: {
  search: string
  status: StatusFilter
  region: string
  sort: SortKey
  regions: string[]
  onSearch: (value: string) => void
  onStatus: (value: StatusFilter) => void
  onRegion: (value: string) => void
  onSort: (value: SortKey) => void
}) {
  return (
    <SurfaceCard title="Facility controls" description="Find, filter, and sort the network without leaving the overview.">
      <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5}>
        <TextField
          label="Search facilities"
          value={search}
          onChange={(e) => onSearch(e.target.value)}
          fullWidth
          slotProps={{ htmlInput: { 'aria-label': 'Search facilities' } }}
        />
        <FormControl sx={{ minWidth: { xs: '100%', md: 180 } }}>
          <InputLabel id="network-status-filter-label">Status</InputLabel>
          <Select
            labelId="network-status-filter-label"
            label="Status"
            value={status}
            onChange={(e) => onStatus(e.target.value as StatusFilter)}
          >
            <MenuItem value="all">All statuses</MenuItem>
            <MenuItem value="critical">Critical</MenuItem>
            <MenuItem value="watch">Watch</MenuItem>
            <MenuItem value="good">Good</MenuItem>
          </Select>
        </FormControl>
        <FormControl sx={{ minWidth: { xs: '100%', md: 180 } }}>
          <InputLabel id="network-region-filter-label">Region</InputLabel>
          <Select
            labelId="network-region-filter-label"
            label="Region"
            value={region}
            onChange={(e) => onRegion(e.target.value)}
          >
            <MenuItem value="all">All regions</MenuItem>
            {regions.map((item) => (
              <MenuItem key={item} value={item}>
                {item}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        <FormControl sx={{ minWidth: { xs: '100%', md: 190 } }}>
          <InputLabel id="network-sort-label">Sort by</InputLabel>
          <Select labelId="network-sort-label" label="Sort by" value={sort} onChange={(e) => onSort(e.target.value as SortKey)}>
            <MenuItem value="attention">Attention first</MenuItem>
            <MenuItem value="revenue">Revenue</MenuItem>
            <MenuItem value="occupancy">Occupancy</MenuItem>
            <MenuItem value="patients">Patients seen</MenuItem>
            <MenuItem value="beds">Beds</MenuItem>
            <MenuItem value="region">Region</MenuItem>
            <MenuItem value="name">Name</MenuItem>
          </Select>
        </FormControl>
      </Stack>
    </SurfaceCard>
  )
}

function FacilityCard({ facility }: { facility: Facility }) {
  const metrics = [
    { label: 'Revenue', value: formatRevenue(facility.latest_revenue_pesewas) },
    { label: 'Occupancy', value: formatOccupancy(facility.occupancy_rate) },
    { label: 'Patients', value: facility.patients_seen == null ? 'No reading' : fmt.number(facility.patients_seen) },
    { label: 'Beds', value: fmt.number(facility.beds) },
  ]

  return (
    <Card variant="outlined" component="article" aria-label={facility.name} sx={{ height: '100%' }}>
      <CardActionArea component={RouterLink} to={`/facilities/${facility.id}`}>
        <CardContent sx={{ minHeight: 218 }}>
          <Stack spacing={1.5}>
            <Stack direction="row" spacing={1} sx={{ justifyContent: 'space-between', alignItems: 'flex-start' }}>
              <Typography variant="h6">{facility.name}</Typography>
              <StatusChip status={facility.status as FacilityStatus} />
            </Stack>
            <Typography variant="body2" color="text.secondary">
              {facility.town}, {facility.region}
            </Typography>
            <Box
              sx={{
                display: 'grid',
                gap: 1,
                gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
              }}
            >
              {metrics.map((item) => (
                <Box
                  key={item.label}
                  sx={{
                    border: 1,
                    borderColor: 'divider',
                    borderRadius: 1.5,
                    bgcolor: 'background.default',
                    p: 1,
                    minHeight: 66,
                  }}
                >
                  <Typography variant="caption" color="text.secondary" sx={{ display: 'block', fontWeight: 800 }}>
                    {item.label}
                  </Typography>
                  <Typography variant="body2" sx={{ mt: 0.25, fontWeight: 800, overflowWrap: 'anywhere' }}>
                    {item.value}
                  </Typography>
                </Box>
              ))}
            </Box>
          </Stack>
        </CardContent>
      </CardActionArea>
    </Card>
  )
}

/** NetworkOverview renders the network summary and worst-first facility grid. */
export function NetworkOverview({ facilities }: { facilities: Facility[] }) {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<StatusFilter>('all')
  const [region, setRegion] = useState('all')
  const [sort, setSort] = useState<SortKey>('attention')
  const resetKey = `${search}|${status}|${region}|${sort}`
  const regions = useMemo(() => [...new Set(facilities.map((f) => f.region))].sort(), [facilities])
  const visible = useMemo(() => {
    const query = search.trim().toLowerCase()
    const filtered = facilities.filter((facility) => {
      const matchesQuery =
        query === '' ||
        facility.name.toLowerCase().includes(query) ||
        facility.town.toLowerCase().includes(query) ||
        facility.region.toLowerCase().includes(query)
      const matchesStatus = status === 'all' || facility.status === status
      const matchesRegion = region === 'all' || facility.region === region
      return matchesQuery && matchesStatus && matchesRegion
    })
    return sortFacilities(filtered, sort)
  }, [facilities, region, search, sort, status])
  const pager = usePagination(visible, { initialPageSize: 9, resetKey })

  return (
    <Stack spacing={3}>
      <NetworkSummary facilities={facilities} />
      <NetworkControls
        search={search}
        status={status}
        region={region}
        sort={sort}
        regions={regions}
        onSearch={setSearch}
        onStatus={setStatus}
        onRegion={setRegion}
        onSort={setSort}
      />
      <Box sx={GRID}>
        {pager.pageItems.map((f) => (
          <FacilityCard key={f.id} facility={f} />
        ))}
      </Box>
      <PaginationControls
        id="network-facilities"
        itemLabel="facilities"
        page={pager.page}
        pageCount={pager.pageCount}
        pageSize={pager.pageSize}
        pageSizeOptions={[9, 18, 36]}
        total={pager.total}
        onPageChange={pager.setPage}
        onPageSizeChange={pager.setPageSize}
      />
      {visible.length === 0 ? (
        <Alert severity="info">No facilities match those controls.</Alert>
      ) : null}
    </Stack>
  )
}

/** NetworkScreen is the single-pane view of the whole facility network. */
export function NetworkScreen() {
  const { data, isLoading, isError } = useFacilities()

  return (
    <Stack spacing={3}>
      <PageHeader
        title="Network"
        eyebrow="Facilities"
        description="A worst-first view of health across the Gigmann estate."
        icon={HubOutlined}
      />
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
