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
import AddTaskOutlined from '@mui/icons-material/AddTaskOutlined'
import DoneAllOutlined from '@mui/icons-material/DoneAllOutlined'
import NotificationsActiveOutlined from '@mui/icons-material/NotificationsActiveOutlined'
import VisibilityOffOutlined from '@mui/icons-material/VisibilityOffOutlined'
import { useState } from 'react'

import { useAlerts, useUpdateAlertStatus, type AlertItem } from '../api/useAlerts'
import { useCreateTask } from '../api/useTasks'
import { EmptyState } from './EmptyState'
import { PaginationControls, usePagination } from './PaginationControls'
import { StatusChip, type FacilityStatus } from './StatusChip'
import { SurfaceCard } from './SurfaceCard'

function priorityFor(severity: FacilityStatus) {
  return severity === 'critical' ? 'high' : severity === 'watch' ? 'medium' : 'low'
}

function AlertCard({
  alert,
  onResolve,
  onDismiss,
  onTask,
  busy,
}: {
  alert: AlertItem
  onResolve: (alert: AlertItem) => void
  onDismiss: (alert: AlertItem) => void
  onTask: (alert: AlertItem) => void
  busy: boolean
}) {
  return (
    <Card variant="outlined" component="article" aria-label={alert.title}>
      <CardContent>
        <Stack spacing={1.5}>
          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} sx={{ justifyContent: 'space-between', gap: 1 }}>
            <Stack direction="row" spacing={1} sx={{ alignItems: 'center', flexWrap: 'wrap', gap: 1 }}>
              <StatusChip status={alert.severity as FacilityStatus} />
              <Chip size="small" variant="outlined" label={alert.type.replaceAll('_', ' ')} sx={{ textTransform: 'capitalize' }} />
              {alert.facility_id ? <Chip size="small" variant="outlined" label={alert.facility_id} /> : null}
            </Stack>
            <Typography variant="caption" color="text.secondary" sx={{ fontWeight: 700, textTransform: 'uppercase' }}>
              {alert.status}
            </Typography>
          </Stack>
          <Box>
            <Typography variant="body1" sx={{ fontWeight: 800 }}>
              {alert.title}
            </Typography>
            {alert.detail ? (
              <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5, lineHeight: 1.6 }}>
                {alert.detail}
              </Typography>
            ) : null}
          </Box>
          <Stack direction="row" spacing={1} sx={{ flexWrap: 'wrap', gap: 1 }}>
            <Button size="small" variant="outlined" startIcon={<AddTaskOutlined />} onClick={() => onTask(alert)} disabled={busy}>
              Turn into task
            </Button>
            <Button size="small" startIcon={<DoneAllOutlined />} onClick={() => onResolve(alert)} disabled={busy}>
              Resolve
            </Button>
            <Button size="small" color="inherit" startIcon={<VisibilityOffOutlined />} onClick={() => onDismiss(alert)} disabled={busy}>
              Dismiss
            </Button>
          </Stack>
        </Stack>
      </CardContent>
    </Card>
  )
}

/** AttentionFeed exposes the ranked alert feed and lets leaders act without leaving the page. */
export function AttentionFeed({ limit = 20, compact = false, pageSize = compact ? 3 : 6 }: { limit?: number; compact?: boolean; pageSize?: number }) {
  const { data, isLoading, isError } = useAlerts(limit)
  const updateStatus = useUpdateAlertStatus()
  const createTask = useCreateTask()
  const [taskAdded, setTaskAdded] = useState(false)
  const alerts = data?.alerts ?? []
  const pager = usePagination(alerts, { initialPageSize: pageSize })
  const busy = updateStatus.isPending || createTask.isPending

  const turnIntoTask = (alert: AlertItem) => {
    createTask.mutate(
      {
        title: alert.title,
        detail: alert.detail,
        facility_id: alert.facility_id,
        priority: priorityFor(alert.severity as FacilityStatus),
        source: 'alert',
      },
      { onSuccess: () => setTaskAdded(true) },
    )
  }

  return (
    <SurfaceCard
      title="Attention feed"
      description={compact ? undefined : 'Open network alerts ranked worst-first, with task and status actions.'}
      icon={NotificationsActiveOutlined}
      sx={compact ? { height: '100%' } : undefined}
    >
      {isLoading ? (
        <Stack spacing={1.5} data-testid="attention-feed-skeleton">
          {[0, 1, 2].map((i) => (
            <Skeleton key={i} variant="rounded" height={compact ? 92 : 126} />
          ))}
        </Stack>
      ) : isError ? (
        <Alert severity="error">Couldn&apos;t load the attention feed. Try again shortly.</Alert>
      ) : alerts.length === 0 ? (
        <EmptyState
          compact={compact}
          icon={NotificationsActiveOutlined}
          title="No open alerts"
          description="The network attention feed is clear."
        />
      ) : (
        <Stack spacing={1.5}>
          {pager.pageItems.map((alert) => (
            <AlertCard
              key={alert.id}
              alert={alert}
              busy={busy}
              onResolve={(item) => updateStatus.mutate({ id: item.id, status: 'resolved' })}
              onDismiss={(item) => updateStatus.mutate({ id: item.id, status: 'dismissed' })}
              onTask={turnIntoTask}
            />
          ))}
          <PaginationControls
            id="attention-feed"
            itemLabel="alerts"
            page={pager.page}
            pageCount={pager.pageCount}
            pageSize={pager.pageSize}
            pageSizeOptions={compact ? [3, 6, 12] : [6, 12, 20]}
            total={pager.total}
            onPageChange={pager.setPage}
            onPageSizeChange={pager.setPageSize}
          />
        </Stack>
      )}
      <Snackbar
        open={taskAdded}
        autoHideDuration={3000}
        onClose={() => setTaskAdded(false)}
        message="Added to My Day"
      />
      <Snackbar
        open={createTask.isError}
        autoHideDuration={4000}
        onClose={() => createTask.reset()}
        message="Couldn&apos;t add the alert to My Day."
      />
    </SurfaceCard>
  )
}
