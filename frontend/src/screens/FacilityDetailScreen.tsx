import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Chip from '@mui/material/Chip'
import Skeleton from '@mui/material/Skeleton'
import Snackbar from '@mui/material/Snackbar'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import useMediaQuery from '@mui/material/useMediaQuery'
import ArrowBackOutlined from '@mui/icons-material/ArrowBackOutlined'
import ContentCopyOutlined from '@mui/icons-material/ContentCopyOutlined'
import ForumOutlined from '@mui/icons-material/ForumOutlined'
import Inventory2Outlined from '@mui/icons-material/Inventory2Outlined'
import LocalHospitalOutlined from '@mui/icons-material/LocalHospitalOutlined'
import NotificationsActiveOutlined from '@mui/icons-material/NotificationsActiveOutlined'
import PeopleAltOutlined from '@mui/icons-material/PeopleAltOutlined'
import SendOutlined from '@mui/icons-material/SendOutlined'
import TaskAltOutlined from '@mui/icons-material/TaskAltOutlined'
import TrendingUpOutlined from '@mui/icons-material/TrendingUpOutlined'
import type { SvgIconComponent } from '@mui/icons-material'
import { Link as RouterLink, useParams } from 'react-router-dom'
import { useState } from 'react'

import { useCreateDraft } from '../api/useDrafts'
import { useFacilityDetail, type FacilityDetail } from '../api/useFacilityDetail'
import { useCreateTask } from '../api/useTasks'
import { EmptyState } from '../components/EmptyState'
import { PageHeader } from '../components/PageHeader'
import { PaginationControls, usePagination } from '../components/PaginationControls'
import { StatusChip, type FacilityStatus } from '../components/StatusChip'
import { SurfaceCard } from '../components/SurfaceCard'
import { KpiCard } from './KpisScreen'

const KPI_GRID = {
  display: 'grid',
  gap: 2,
  gridTemplateColumns: { xs: '1fr', lg: 'repeat(2, 1fr)' },
} as const

function Section({
  title,
  count,
  icon,
  children,
}: {
  title: string
  count: number
  icon: SvgIconComponent
  children: React.ReactNode
}) {
  return (
    <SurfaceCard title={`${title} (${count})`} icon={icon}>
      {count === 0 ? (
        <EmptyState compact icon={icon} title={`No ${title.toLowerCase()}`} description="Nothing needs attention in this section." />
      ) : (
        children
      )}
    </SurfaceCard>
  )
}

function priorityFor(status: FacilityStatus) {
  return status === 'critical' ? 'high' : status === 'watch' ? 'medium' : 'low'
}

function Detail({ data }: { data: FacilityDetail }) {
  const f = data.facility
  const reduceMotion = useMediaQuery('(prefers-reduced-motion: reduce)')
  const createTask = useCreateTask()
  const createDraft = useCreateDraft()
  const [taskAdded, setTaskAdded] = useState(false)
  const [draftCopied, setDraftCopied] = useState(false)
  const kpis = data.kpis ?? []
  const kpiPager = usePagination(kpis, { initialPageSize: 4 })
  const alertPager = usePagination(data.alerts, { initialPageSize: 4 })
  const inventoryPager = usePagination(data.inventory, { initialPageSize: 5 })
  const staffPager = usePagination(data.staff, { initialPageSize: 5 })

  const makeTask = () => {
    createTask.mutate(
      {
        title: `Review ${f.name} posture`,
        detail: `Follow up on ${f.name} (${f.status}) using current facility detail, alerts, staffing, and inventory.`,
        facility_id: f.id,
        priority: priorityFor(f.status as FacilityStatus),
        source: 'manual',
      },
      { onSuccess: () => setTaskAdded(true) },
    )
  }

  const generateManagerDraft = () => {
    createDraft.mutate({
      kind: 'message',
      facility_id: f.id,
      instruction: `Draft a concise manager follow-up for ${f.name} based on its current alerts, inventory, staffing, and KPI trend.`,
    })
  }

  const copyDraft = async () => {
    if (createDraft.data?.draft) {
      await navigator.clipboard?.writeText(createDraft.data.draft)
      setDraftCopied(true)
    }
  }

  return (
    <Stack spacing={3}>
      <PageHeader
        title={f.name}
        eyebrow="Facility detail"
        description={`${f.town}, ${f.region} · ${f.beds} beds`}
        icon={LocalHospitalOutlined}
        actions={<StatusChip status={f.status as FacilityStatus} />}
      />

      <SurfaceCard
        title="Quick actions"
        description="Move from observation to follow-up without losing the facility context."
        icon={TaskAltOutlined}
      >
        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} sx={{ flexWrap: 'wrap', gap: 1 }}>
          <Button variant="contained" startIcon={<TaskAltOutlined />} onClick={makeTask} disabled={createTask.isPending}>
            Add follow-up task
          </Button>
          <Button
            variant="outlined"
            startIcon={<ForumOutlined />}
            component={RouterLink}
            to="/ask"
            state={{ question: `What should I do about ${f.name} today?` }}
          >
            Ask about this facility
          </Button>
          <Button variant="outlined" startIcon={<SendOutlined />} onClick={generateManagerDraft} disabled={createDraft.isPending}>
            Draft manager message
          </Button>
        </Stack>
        {createTask.isError ? (
          <Alert severity="error" sx={{ mt: 2 }}>
            Couldn&apos;t create the follow-up task. Try again shortly.
          </Alert>
        ) : null}
        {createDraft.isError ? (
          <Alert severity="error" sx={{ mt: 2 }}>
            Couldn&apos;t generate the draft. Try again shortly.
          </Alert>
        ) : null}
      </SurfaceCard>

      {createDraft.data ? (
        <SurfaceCard
          title="Generated draft"
          description="Draft-only output. Nothing is sent until an executive copies and sends it."
          icon={SendOutlined}
          actions={
            <Button size="small" startIcon={<ContentCopyOutlined />} onClick={() => void copyDraft()}>
              Copy
            </Button>
          }
        >
          <Typography variant="body1" sx={{ whiteSpace: 'pre-wrap', lineHeight: 1.8 }}>
            {createDraft.data.draft}
          </Typography>
        </SurfaceCard>
      ) : null}

      <Section title="KPI trends" count={kpis.length} icon={TrendingUpOutlined}>
        <Stack spacing={2}>
          <Box sx={KPI_GRID}>
            {kpiPager.pageItems.map((kpi) => (
              <KpiCard key={kpi.key} kpi={kpi} reduceMotion={reduceMotion} />
            ))}
          </Box>
          <PaginationControls
            id="facility-kpis"
            itemLabel="KPIs"
            page={kpiPager.page}
            pageCount={kpiPager.pageCount}
            pageSize={kpiPager.pageSize}
            pageSizeOptions={[4, 8, 12]}
            total={kpiPager.total}
            onPageChange={kpiPager.setPage}
            onPageSizeChange={kpiPager.setPageSize}
          />
        </Stack>
      </Section>

      <Section title="Alerts" count={data.alerts.length} icon={NotificationsActiveOutlined}>
        <Stack spacing={1}>
          {alertPager.pageItems.map((a) => (
            <Card key={a.id} variant="outlined">
              <CardContent>
                <Stack spacing={0.75}>
                  <Stack direction="row" spacing={1} sx={{ alignItems: 'center', flexWrap: 'wrap', gap: 1 }}>
                    <StatusChip status={a.severity as FacilityStatus} />
                    <Typography variant="body2" sx={{ fontWeight: 700 }}>
                      {a.title}
                    </Typography>
                  </Stack>
                  {a.detail ? (
                    <Typography variant="body2" color="text.secondary">
                      {a.detail}
                    </Typography>
                  ) : null}
                </Stack>
              </CardContent>
            </Card>
          ))}
          <PaginationControls
            id="facility-alerts"
            itemLabel="alerts"
            page={alertPager.page}
            pageCount={alertPager.pageCount}
            pageSize={alertPager.pageSize}
            pageSizeOptions={[4, 8, 12]}
            total={alertPager.total}
            onPageChange={alertPager.setPage}
            onPageSizeChange={alertPager.setPageSize}
          />
        </Stack>
      </Section>

      <Section title="Inventory" count={data.inventory.length} icon={Inventory2Outlined}>
        <Stack spacing={1}>
          {inventoryPager.pageItems.map((it) => (
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
          <PaginationControls
            id="facility-inventory"
            itemLabel="items"
            page={inventoryPager.page}
            pageCount={inventoryPager.pageCount}
            pageSize={inventoryPager.pageSize}
            pageSizeOptions={[5, 10, 20]}
            total={inventoryPager.total}
            onPageChange={inventoryPager.setPage}
            onPageSizeChange={inventoryPager.setPageSize}
          />
        </Stack>
      </Section>

      <Section title="Staff" count={data.staff.length} icon={PeopleAltOutlined}>
        <Stack spacing={1}>
          {staffPager.pageItems.map((m) => (
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
          <PaginationControls
            id="facility-staff"
            itemLabel="staff"
            page={staffPager.page}
            pageCount={staffPager.pageCount}
            pageSize={staffPager.pageSize}
            pageSizeOptions={[5, 10, 20]}
            total={staffPager.total}
            onPageChange={staffPager.setPage}
            onPageSizeChange={staffPager.setPageSize}
          />
        </Stack>
      </Section>

      <Snackbar
        open={taskAdded}
        autoHideDuration={3000}
        onClose={() => setTaskAdded(false)}
        message="Added to My Day"
      />
      <Snackbar
        open={draftCopied}
        autoHideDuration={2500}
        onClose={() => setDraftCopied(false)}
        message="Draft copied"
      />
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
        <ArrowBackOutlined fontSize="small" sx={{ mr: 0.75 }} />
        Network
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
