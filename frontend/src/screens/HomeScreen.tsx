import Button from '@mui/material/Button'
import Snackbar from '@mui/material/Snackbar'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { useBrief, type Brief } from '../api/useBrief'
import { useCreateTask } from '../api/useTasks'
import { DailyBrief } from '../components/DailyBrief'
import { briefToMarkdown } from './exportBrief'

/** HomeScreen is the hero "Today" view — the AI-narrated Daily Brief. */
export function HomeScreen() {
  const { data, isLoading, isError } = useBrief()
  const navigate = useNavigate()
  const createTask = useCreateTask()
  const [taskAdded, setTaskAdded] = useState(false)

  const onTask = (item: Brief['items'][number]) => {
    const priority = item.severity === 'critical' ? 'high' : item.severity === 'watch' ? 'medium' : 'low'
    createTask.mutate(
      { title: item.headline, facility_id: item.facility_id, priority, source: 'brief' },
      { onSuccess: () => setTaskAdded(true) },
    )
  }

  // A suggested action on a brief item jumps to Ask with the question prefilled.
  const askAbout = (action: string, facilityId: string) => {
    navigate('/ask', { state: { question: `${action} — ${facilityId}` } })
  }

  const copy = () => {
    if (data) {
      void navigator.clipboard?.writeText(briefToMarkdown(data))
    }
  }
  const download = () => {
    if (!data) {
      return
    }
    const url = URL.createObjectURL(new Blob([briefToMarkdown(data)], { type: 'text/markdown' }))
    const link = document.createElement('a')
    link.href = url
    link.download = `daily-brief-${data.date}.md`
    link.click()
    URL.revokeObjectURL(url)
  }

  return (
    <Stack spacing={3}>
      <Stack direction="row" spacing={1} sx={{ justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap' }}>
        <Typography variant="h1" sx={{ fontSize: { xs: '2rem', md: '2.5rem' } }}>
          The Brief
        </Typography>
        {data ? (
          <Stack direction="row" spacing={1}>
            <Button size="small" variant="outlined" onClick={copy}>
              Copy
            </Button>
            <Button size="small" variant="outlined" onClick={download}>
              Download
            </Button>
          </Stack>
        ) : null}
      </Stack>
      <DailyBrief brief={data} isLoading={isLoading} isError={isError} onAction={askAbout} onTask={onTask} />
      <Snackbar
        open={taskAdded}
        autoHideDuration={3000}
        onClose={() => setTaskAdded(false)}
        message="Added to My Day"
      />
    </Stack>
  )
}
