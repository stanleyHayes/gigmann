import Alert from '@mui/material/Alert'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Chip from '@mui/material/Chip'
import Skeleton from '@mui/material/Skeleton'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import { Link as RouterLink, useParams } from 'react-router-dom'

import { useFacilityDetail, type FacilityDetail } from '../api/useFacilityDetail'
import { StatusChip, type FacilityStatus } from '../components/StatusChip'

function Section({ title, count, children }: { title: string; count: number; children: React.ReactNode }) {
  return (
    <Stack spacing={1}>
      <Typography variant="h6" sx={{ fontWeight: 600 }}>
        {title} ({count})
      </Typography>
      {count === 0 ? (
        <Typography variant="body2" color="text.secondary">
          Nothing here.
        </Typography>
      ) : (
        children
      )}
    </Stack>
  )
}

function Detail({ data }: { data: FacilityDetail }) {
  const f = data.facility
  return (
    <Stack spacing={3}>
      <Stack spacing={1}>
        <Stack direction="row" spacing={1} sx={{ justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <Typography variant="h1" sx={{ fontSize: { xs: '1.75rem', md: '2.25rem' } }}>
            {f.name}
          </Typography>
          <StatusChip status={f.status as FacilityStatus} />
        </Stack>
        <Typography variant="body2" color="text.secondary">
          {f.town}, {f.region} · {f.beds} beds
        </Typography>
      </Stack>

      <Section title="Alerts" count={data.alerts.length}>
        <Stack spacing={1}>
          {data.alerts.map((a) => (
            <Card key={a.id} variant="outlined">
              <CardContent>
                <Stack direction="row" spacing={1} sx={{ alignItems: 'center' }}>
                  <StatusChip status={a.severity as FacilityStatus} />
                  <Typography variant="body2" sx={{ fontWeight: 600 }}>
                    {a.title}
                  </Typography>
                </Stack>
              </CardContent>
            </Card>
          ))}
        </Stack>
      </Section>

      <Section title="Inventory" count={data.inventory.length}>
        <Stack spacing={1}>
          {data.inventory.map((it) => (
            <Card key={it.id} variant="outlined">
              <CardContent>
                <Stack direction="row" spacing={1} sx={{ justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap' }}>
                  <Typography variant="body2" sx={{ fontWeight: 600 }}>
                    {it.name}
                  </Typography>
                  <Stack direction="row" spacing={1}>
                    <Chip size="small" variant="outlined" label={`${it.stock_level} in stock`} />
                    {it.days_of_stock != null ? (
                      <Chip size="small" variant="outlined" label={`~${Math.round(it.days_of_stock)}d left`} />
                    ) : null}
                    {it.stockout_imminent ? <Chip size="small" color="error" label="Stockout imminent" /> : null}
                  </Stack>
                </Stack>
              </CardContent>
            </Card>
          ))}
        </Stack>
      </Section>

      <Section title="Staff" count={data.staff.length}>
        <Stack spacing={1}>
          {data.staff.map((m) => (
            <Card key={m.id} variant="outlined">
              <CardContent>
                <Stack direction="row" spacing={1} sx={{ justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap' }}>
                  <Typography variant="body2">
                    <strong>{m.role}</strong> — {m.name}
                  </Typography>
                  <Stack direction="row" spacing={1}>
                    {m.attrition_risk >= 0.6 ? <Chip size="small" color="warning" label="Attrition risk" /> : null}
                    {m.licence_expiry ? <Chip size="small" variant="outlined" label={`Licence ${m.licence_expiry}`} /> : null}
                  </Stack>
                </Stack>
              </CardContent>
            </Card>
          ))}
        </Stack>
      </Section>
    </Stack>
  )
}

/** FacilityDetailScreen is the drill-down reached from a Network card. */
export function FacilityDetailScreen() {
  const { facilityId = '' } = useParams()
  const { data, isLoading, isError } = useFacilityDetail(facilityId)

  return (
    <Stack spacing={3}>
      <Button component={RouterLink} to="/network" size="small" sx={{ alignSelf: 'flex-start' }}>
        ← Network
      </Button>
      {isLoading ? (
        <Stack spacing={2} data-testid="facility-skeleton">
          <Skeleton variant="text" width="50%" height={40} />
          <Skeleton variant="rounded" height={88} />
          <Skeleton variant="rounded" height={88} />
        </Stack>
      ) : isError || !data ? (
        <Alert severity="error">Couldn&apos;t load this facility.</Alert>
      ) : (
        <Detail data={data} />
      )}
    </Stack>
  )
}
