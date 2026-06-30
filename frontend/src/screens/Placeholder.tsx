import Stack from '@mui/material/Stack'
import SearchOffOutlined from '@mui/icons-material/SearchOffOutlined'

import { EmptyState } from '../components/EmptyState'
import { PageHeader } from '../components/PageHeader'

/** Placeholder is a lightweight stand-in for screens planned in later stories. */
export function Placeholder({ title, note }: { title: string; note?: string }) {
  return (
    <Stack spacing={2}>
      <PageHeader title={title} eyebrow="Cockpit" />
      <EmptyState
        icon={SearchOffOutlined}
        title="Page unavailable"
        description={note ?? 'This screen is planned in a later story.'}
      />
    </Stack>
  )
}
