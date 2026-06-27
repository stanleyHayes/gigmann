import Alert from '@mui/material/Alert'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Chip from '@mui/material/Chip'
import Skeleton from '@mui/material/Skeleton'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'

import { useTasks, type Task } from '../api/useTasks'
import { useAuth } from '../auth/authContext'

const STATUS_LABEL: Record<string, string> = { todo: 'To do', in_progress: 'In progress', done: 'Done' }

function isStalled(t: Task): boolean {
  return Boolean(t.due_date) && t.status !== 'done' && new Date(t.due_date as string) < new Date()
}

/** DelegationScreen shows actions assigned to others and their follow-through. */
export function DelegationScreen() {
  const { data: tasks = [], isLoading, isError } = useTasks()
  const { user } = useAuth()

  const delegated = tasks.filter((t) => t.assigned_to && t.assigned_to !== user?.name)
  const byAssignee = new Map<string, Task[]>()
  for (const t of delegated) {
    const key = t.assigned_to as string
    byAssignee.set(key, [...(byAssignee.get(key) ?? []), t])
  }

  return (
    <Stack spacing={3}>
      <Typography variant="h1" sx={{ fontSize: { xs: '2rem', md: '2.5rem' } }}>
        Delegation
      </Typography>
      <Typography variant="body2" color="text.secondary">
        Actions you&apos;ve assigned to facility managers — and whether they&apos;re moving.
      </Typography>

      {isLoading ? (
        <Skeleton data-testid="delegation-skeleton" variant="rounded" height={120} />
      ) : isError ? (
        <Alert severity="error">Couldn&apos;t load delegated work.</Alert>
      ) : byAssignee.size === 0 ? (
        <Typography variant="body2" color="text.secondary">Nothing delegated right now.</Typography>
      ) : (
        [...byAssignee.entries()].map(([assignee, items]) => (
          <Card key={assignee} variant="outlined">
            <CardContent>
              <Stack spacing={1.5}>
                <Typography variant="h6" sx={{ fontWeight: 600 }}>{assignee}</Typography>
                {items.map((t) => (
                  <Stack key={t.id} direction="row" spacing={1} sx={{ alignItems: 'center', flexWrap: 'wrap', gap: 1 }}>
                    <Typography variant="body2" sx={{ flexGrow: 1 }}>{t.title}</Typography>
                    <Chip size="small" label={STATUS_LABEL[t.status] ?? t.status} />
                    {isStalled(t) ? <Chip size="small" color="warning" label="Stalled" /> : null}
                  </Stack>
                ))}
              </Stack>
            </CardContent>
          </Card>
        ))
      )}
    </Stack>
  )
}
