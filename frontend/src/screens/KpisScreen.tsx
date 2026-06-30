import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Skeleton from '@mui/material/Skeleton'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import useMediaQuery from '@mui/material/useMediaQuery'
import { useTheme } from '@mui/material/styles'
import { LineChart } from '@mui/x-charts/LineChart'
import InsightsOutlined from '@mui/icons-material/InsightsOutlined'

import { useMetrics, type Kpi } from '../api/useMetrics'
import { fmt } from '../i18n/locale'
import { monoFont, statusColors } from '../theme'
import { PageHeader } from '../components/PageHeader'
import { PaginationControls, usePagination } from '../components/PaginationControls'

const GRID = {
  display: 'grid',
  gap: 2,
  gridTemplateColumns: { xs: '1fr', md: 'repeat(2, 1fr)' },
} as const

const ARROW: Record<Kpi['direction'], string> = { up: '↑', down: '↓', flat: '→' }
const TONE_WORD = { good: 'improved', bad: 'worsened', flat: 'unchanged' } as const

type Tone = 'good' | 'bad' | 'flat'

function toneOf(kpi: Kpi): Tone {
  if (kpi.direction === 'flat') {
    return 'flat'
  }
  return (kpi.direction === 'up') === kpi.higher_is_better ? 'good' : 'bad'
}

function formatValue(unit: Kpi['unit'], value: number): string {
  switch (unit) {
    case 'pesewas':
      // Format from minor units without losing pesewas (the engine owns the figure).
      return fmt.cedis(value / 100)
    case 'ratio':
      return `${(value * 100).toFixed(1)}%`
    default:
      return Math.round(value).toLocaleString('en-US')
  }
}

function formatDelta(deltaPct: number): string {
  return `${Math.abs(Math.round(deltaPct * 100))}%`
}

// shortDate turns an ISO date ("2026-06-12") into a compact "6/12" tick label
// without constructing a Date (avoids timezone drift on the axis).
function shortDate(iso: string): string {
  const [, m, d] = iso.split('-')
  return `${Number(m)}/${Number(d)}`
}

/** KpiCard shows one KPI's current value, week-over-week delta, and daily trend. */
export function KpiCard({ kpi, reduceMotion }: { kpi: Kpi; reduceMotion: boolean }) {
  const theme = useTheme()
  const tone = toneOf(kpi)
  const deltaColor =
    tone === 'good' ? statusColors.good : tone === 'bad' ? statusColors.critical : theme.palette.text.secondary
  const scale = (v: number) => (kpi.unit === 'pesewas' ? v / 100 : kpi.unit === 'ratio' ? v * 100 : v)

  return (
    <Card variant="outlined" sx={{ height: '100%', overflow: 'hidden', borderLeft: 4, borderLeftColor: deltaColor }}>
      <CardContent>
        <Stack spacing={1.25}>
          <Typography variant="overline" color="text.secondary" sx={{ fontWeight: 800, letterSpacing: 0 }}>
            {kpi.label}
          </Typography>
          <Stack direction="row" spacing={1} sx={{ alignItems: 'baseline' }}>
            <Typography variant="h5" sx={{ fontFamily: monoFont, fontWeight: 600 }}>
              {formatValue(kpi.unit, kpi.current)}
            </Typography>
            <Typography variant="body2" sx={{ color: deltaColor, fontWeight: 600 }} aria-label={TONE_WORD[tone]}>
              {ARROW[kpi.direction]} {formatDelta(kpi.delta_pct)}
            </Typography>
          </Stack>
          <LineChart
            height={170}
            skipAnimation={reduceMotion}
            hideLegend
            grid={{ horizontal: true }}
            margin={{ left: 52, right: 12, top: 8, bottom: 24 }}
            series={[
              {
                data: kpi.series.map((p) => scale(p.value)),
                color: theme.palette.primary.main,
                curve: 'monotoneX',
                showMark: false,
              },
            ]}
            xAxis={[{ data: kpi.series.map((p) => shortDate(p.date)), scaleType: 'point' }]}
          />
        </Stack>
      </CardContent>
    </Card>
  )
}

/** KpisScreen is the executive KPI dashboard backed by GET /api/v1/metrics. */
export function KpisScreen() {
  const reduceMotion = useMediaQuery('(prefers-reduced-motion: reduce)')
  const { data, isLoading, isError } = useMetrics()
  const kpis = data?.kpis ?? []
  const pager = usePagination(kpis, { initialPageSize: 4 })

  return (
    <Stack spacing={3}>
      <PageHeader
        title="Executive KPIs"
        eyebrow="Network pulse"
        description="Deterministic financial, demand, occupancy, and denial trends."
        icon={InsightsOutlined}
      />
      {isLoading ? (
        <Box sx={GRID} data-testid="kpis-skeleton">
          {[0, 1, 2, 3].map((i) => (
            <Skeleton key={i} variant="rounded" height={232} />
          ))}
        </Box>
      ) : isError || !data ? (
        <Alert severity="error">Couldn&apos;t load the KPIs. Try again shortly.</Alert>
      ) : (
        <Stack spacing={2}>
          <Box sx={GRID}>
            {pager.pageItems.map((k) => (
              <KpiCard key={k.key} kpi={k} reduceMotion={reduceMotion} />
            ))}
          </Box>
          <PaginationControls
            id="kpis"
            itemLabel="KPIs"
            page={pager.page}
            pageCount={pager.pageCount}
            pageSize={pager.pageSize}
            pageSizeOptions={[4, 8, 12]}
            total={pager.total}
            onPageChange={pager.setPage}
            onPageSizeChange={pager.setPageSize}
          />
        </Stack>
      )}
    </Stack>
  )
}
