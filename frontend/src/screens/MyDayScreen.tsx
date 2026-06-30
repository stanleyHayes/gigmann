import Alert from '@mui/material/Alert'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Checkbox from '@mui/material/Checkbox'
import Chip from '@mui/material/Chip'
import Skeleton from '@mui/material/Skeleton'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import ChecklistOutlined from '@mui/icons-material/ChecklistOutlined'
import TaskAltOutlined from '@mui/icons-material/TaskAltOutlined'

import { useTasks, useUpdateTaskStatus, type Task } from '../api/useTasks'
import { EmptyState } from '../components/EmptyState'
import { PageHeader } from '../components/PageHeader'
import { PaginationControls, usePagination } from '../components/PaginationControls'

const PRIORITY_COLOR: Record<Task['priority'], 'error' | 'warning' | 'default'> = {
  high: 'error',
  medium: 'warning',
  low: 'default',
}
const PRIORITY_BORDER: Record<Task['priority'], string> = {
  high: 'error.main',
  medium: 'warning.main',
  low: 'divider',
}
const STATUS_RANK: Record<Task['status'], number> = { todo: 0, in_progress: 0, done: 1 }
const PRIORITY_RANK: Record<Task['priority'], number> = { high: 0, medium: 1, low: 2 }

// dueLabel turns "2026-06-26" into "Due 6/26" without timezone drift.
function dueLabel(due: string): string {
  const [, m, d] = due.split('-')
  return `Due ${Number(m)}/${Number(d)}`
}

function TaskCard({ task, onToggle }: { task: Task; onToggle: (t: Task, done: boolean) => void }) {
  const done = task.status === 'done'
  return (
    <Card
      variant="outlined"
      sx={{
        position: 'relative',
        overflow: 'hidden',
        opacity: done ? 0.72 : 1,
        borderLeft: 4,
        borderLeftColor: PRIORITY_BORDER[task.priority],
      }}
    >
      <CardContent>
        <Stack direction="row" spacing={1.5} sx={{ alignItems: 'flex-start' }}>
          <Checkbox
            checked={done}
            onChange={(e) => onToggle(task, e.target.checked)}
            slotProps={{ input: { 'aria-label': `Mark “${task.title}” done` } }}
            sx={{ mt: -0.5 }}
          />
          <Stack spacing={0.5} sx={{ flex: 1 }}>
            <Typography
              variant="subtitle1"
              sx={{
                fontWeight: 600,
                textDecoration: done ? 'line-through' : 'none',
                color: done ? 'text.disabled' : 'text.primary',
              }}
            >
              {task.title}
            </Typography>
            {task.detail ? (
              <Typography variant="body2" color="text.secondary">
                {task.detail}
              </Typography>
            ) : null}
            <Stack direction="row" spacing={1} sx={{ flexWrap: 'wrap', gap: 1, pt: 0.5 }}>
              <Chip size="small" label={task.priority} color={PRIORITY_COLOR[task.priority]} sx={{ textTransform: 'capitalize' }} />
              {task.status === 'in_progress' ? <Chip size="small" variant="outlined" label="In progress" /> : null}
              {task.facility_id ? <Chip size="small" variant="outlined" label={task.facility_id} /> : null}
              {task.due_date ? <Chip size="small" variant="outlined" label={dueLabel(task.due_date)} /> : null}
            </Stack>
          </Stack>
        </Stack>
      </CardContent>
    </Card>
  )
}

/** MyDayScreen is the executive's task list: active first, completed sink to the bottom. */
export function MyDayScreen() {
  const { data, isLoading, isError } = useTasks()
  const update = useUpdateTaskStatus()

  const onToggle = (task: Task, done: boolean) => {
    update.mutate({ id: task.id, status: done ? 'done' : 'todo' })
  }

  const sorted = data
    ? [...data].sort(
        (a, b) => STATUS_RANK[a.status] - STATUS_RANK[b.status] || PRIORITY_RANK[a.priority] - PRIORITY_RANK[b.priority],
      )
    : []
  const pager = usePagination(sorted, { initialPageSize: 6 })

  return (
    <Stack spacing={3}>
      <PageHeader
        title="My Day"
        eyebrow="Execution"
        description="Active work first, completed work settled out of the way."
        icon={ChecklistOutlined}
      />
      {isLoading ? (
        <Stack spacing={2} data-testid="myday-skeleton">
          {[0, 1, 2].map((i) => (
            <Skeleton key={i} variant="rounded" height={96} />
          ))}
        </Stack>
      ) : isError || !data ? (
        <Alert severity="error">Couldn&apos;t load your tasks. Try again shortly.</Alert>
      ) : data.length === 0 ? (
        <EmptyState
          icon={TaskAltOutlined}
          title="Nothing on your list today"
          description="Tasks created from brief items and delegated actions will collect here."
        />
      ) : (
        <Stack spacing={2}>
          {pager.pageItems.map((t) => (
            <TaskCard key={t.id} task={t} onToggle={onToggle} />
          ))}
          <PaginationControls
            id="my-day-tasks"
            itemLabel="tasks"
            page={pager.page}
            pageCount={pager.pageCount}
            pageSize={pager.pageSize}
            pageSizeOptions={[6, 12, 24]}
            total={pager.total}
            onPageChange={pager.setPage}
            onPageSizeChange={pager.setPageSize}
          />
        </Stack>
      )}
    </Stack>
  )
}
