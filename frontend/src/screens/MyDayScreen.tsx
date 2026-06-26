import Alert from '@mui/material/Alert'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Checkbox from '@mui/material/Checkbox'
import Chip from '@mui/material/Chip'
import Skeleton from '@mui/material/Skeleton'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'

import { useTasks, useUpdateTaskStatus, type Task } from '../api/useTasks'

const PRIORITY_COLOR: Record<Task['priority'], 'error' | 'warning' | 'default'> = {
  high: 'error',
  medium: 'warning',
  low: 'default',
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
    <Card variant="outlined">
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

  return (
    <Stack spacing={3}>
      <Typography variant="h1" sx={{ fontSize: { xs: '2rem', md: '2.5rem' } }}>
        My Day
      </Typography>
      {isLoading ? (
        <Stack spacing={2} data-testid="myday-skeleton">
          {[0, 1, 2].map((i) => (
            <Skeleton key={i} variant="rounded" height={96} />
          ))}
        </Stack>
      ) : isError || !data ? (
        <Alert severity="error">Couldn&apos;t load your tasks. Try again shortly.</Alert>
      ) : data.length === 0 ? (
        <Alert severity="info">Nothing on your list today.</Alert>
      ) : (
        <Stack spacing={2}>
          {sorted.map((t) => (
            <TaskCard key={t.id} task={t} onToggle={onToggle} />
          ))}
        </Stack>
      )}
    </Stack>
  )
}
