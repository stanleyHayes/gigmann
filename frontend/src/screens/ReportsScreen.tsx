import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import DownloadOutlined from '@mui/icons-material/DownloadOutlined'

import { useBrief } from '../api/useBrief'
import { useMetrics } from '../api/useMetrics'
import { downloadFile, networkReportMarkdown } from './exportBrief'

/** ReportsScreen generates a shareable network report (brief + KPIs) to download. */
export function ReportsScreen() {
  const brief = useBrief()
  const metrics = useMetrics()
  const ready = brief.data !== undefined

  const onDownload = () => {
    if (!brief.data) {
      return
    }
    downloadFile(`network-report-${brief.data.date}.md`, networkReportMarkdown(brief.data, metrics.data))
  }

  return (
    <Stack spacing={3}>
      <Typography variant="h1" sx={{ fontSize: { xs: '2rem', md: '2.5rem' } }}>
        Reports
      </Typography>
      <Typography variant="body2" color="text.secondary">
        Generate a shareable network report — the Daily Brief plus the network KPIs.
      </Typography>
      <Card variant="outlined">
        <CardContent>
          <Stack spacing={2}>
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              Network report
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {ready
                ? `Based on the brief for ${brief.data?.date}.`
                : brief.isError
                  ? 'Couldn’t load the brief — try again shortly.'
                  : 'Preparing the latest figures…'}
            </Typography>
            <Button
              variant="contained"
              startIcon={<DownloadOutlined />}
              disabled={!ready}
              onClick={onDownload}
              sx={{ alignSelf: 'flex-start' }}
            >
              Download report (Markdown)
            </Button>
          </Stack>
        </CardContent>
      </Card>
    </Stack>
  )
}
