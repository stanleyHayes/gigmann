import Alert from '@mui/material/Alert'
import Chip from '@mui/material/Chip'
import Skeleton from '@mui/material/Skeleton'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import AssignmentIndOutlined from '@mui/icons-material/AssignmentIndOutlined'
import GroupsOutlined from '@mui/icons-material/GroupsOutlined'
import { useMemo } from 'react'

import { useTasks, type Task } from '../api/useTasks'
import { useAuth } from '../auth/authContext'
import { EmptyState } from '../components/EmptyState'
import { PageHeader } from '../components/PageHeader'
import { PaginationControls, usePagination } from '../components/PaginationControls'
import { SurfaceCard } from '../components/SurfaceCard'

const STATUS_LABEL: Record<string, string> = { todo: 'To do', in_progress: 'In progress', done: 'Done' }

function isStalled(t: Task): boolean {
  return Boolean(t.due_date) && t.status !== 'done' && new Date(t.due_date as string) < new Date()
}

function slug(value: string): string {
  return value.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)/g, '') || 'assignee'
}

function DelegatedTaskList({ assignee, items }: { assignee: string; items: Task[] }) {
  const pager = usePagination(items, { initialPageSize: 5 })

  return (
    <Stack spacing={1.5}>
      {pager.pageItems.map((t) => (
        <Stack key={t.id} direction="row" spacing={1} sx={{ alignItems: 'center', flexWrap: 'wrap', gap: 1 }}>
          <Typography variant="body2" sx={{ flexGrow: 1 }}>{t.title}</Typography>
          <Chip size="small" label={STATUS_LABEL[t.status] ?? t.status} />
          {isStalled(t) ? <Chip size="small" color="warning" label="Stalled" /> : null}
        </Stack>
      ))}
      <PaginationControls
        id={`delegation-${slug(assignee)}-tasks`}
        itemLabel="tasks"
        page={pager.page}
        pageCount={pager.pageCount}
        pageSize={pager.pageSize}
        pageSizeOptions={[5, 10, 20]}
        total={pager.total}
        onPageChange={pager.setPage}
        onPageSizeChange={pager.setPageSize}
      />
    </Stack>
  )
}

/** DelegationScreen shows actions assigned to others and their follow-through. */
export function DelegationScreen() {
  const { data: tasks = [], isLoading, isError } = useTasks()
  const { user } = useAuth()

  const groups = useMemo(() => {
    const delegated = tasks.filter((t) => t.assigned_to && t.assigned_to !== user?.name)
    const byAssignee = new Map<string, Task[]>()
    for (const t of delegated) {
      const key = t.assigned_to as string
      byAssignee.set(key, [...(byAssignee.get(key) ?? []), t])
    }
    return [...byAssignee.entries()].sort(([a], [b]) => a.localeCompare(b))
  }, [tasks, user?.name])
  const pager = usePagination(groups, { initialPageSize: 4 })

  return (
    <Stack spacing={3}>
      <PageHeader
        title="Delegation"
        eyebrow="Follow-through"
        description="Actions assigned to facility managers and whether they are moving."
        icon={AssignmentIndOutlined}
      />

      {isLoading ? (
        <Skeleton data-testid="delegation-skeleton" variant="rounded" height={120} />
      ) : isError ? (
        <Alert severity="error">Couldn&apos;t load delegated work.</Alert>
      ) : groups.length === 0 ? (
        <EmptyState
          icon={GroupsOutlined}
          title="Nothing delegated right now"
          description="Assigned follow-ups will group by manager once work is handed off."
        />
      ) : (
        <Stack spacing={2}>
          {pager.pageItems.map(([assignee, items]) => (
            <SurfaceCard key={assignee} title={assignee} icon={AssignmentIndOutlined}>
              <DelegatedTaskList assignee={assignee} items={items} />
            </SurfaceCard>
          ))}
          <PaginationControls
            id="delegation-assignees"
            itemLabel="managers"
            page={pager.page}
            pageCount={pager.pageCount}
            pageSize={pager.pageSize}
            pageSizeOptions={[4, 8, 16]}
            total={pager.total}
            onPageChange={pager.setPage}
            onPageSizeChange={pager.setPageSize}
          />
        </Stack>
      )}
    </Stack>
  )
}
