import { useMemo, useRef, useState } from 'react'
import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import DownloadOutlined from '@mui/icons-material/DownloadOutlined'

import { useBrief } from '../api/useBrief'
import { useMetrics } from '../api/useMetrics'
import { chartToPng, downloadFile, downloadPdf, networkReportCsv, networkReportMarkdown } from './exportBrief'

/** ReportsScreen generates shareable network reports (brief + KPIs) to download. */
export function ReportsScreen() {
  const brief = useBrief()
  const metrics = useMetrics()
  const ready = brief.data !== undefined && metrics.data !== undefined
  const previewRef = useRef<HTMLDivElement>(null)
  const [pdfBusy, setPdfBusy] = useState(false)
  const [pdfError, setPdfError] = useState(false)
  const chartUrl = useMemo(() => (metrics.data ? chartToPng(metrics.data) : ''), [metrics.data])

  const onDownloadMarkdown = () => {
    if (!brief.data) {
      return
    }
    downloadFile(`network-report-${brief.data.date}.md`, networkReportMarkdown(brief.data, metrics.data))
  }

  const onDownloadCsv = () => {
    if (!metrics.data) {
      return
    }
    downloadFile(`network-kpis-${metrics.data.as_of}.csv`, networkReportCsv(metrics.data), 'text/csv')
  }

  const onDownloadPdf = async () => {
    if (!previewRef.current || !brief.data) {
      return
    }
    setPdfBusy(true)
    setPdfError(false)
    try {
      await downloadPdf(`network-report-${brief.data.date}.pdf`, previewRef.current)
    } catch {
      setPdfError(true)
    } finally {
      setPdfBusy(false)
    }
  }

  const statusText = () => {
    if (brief.isError || metrics.isError) {
      return "Couldn't load the latest figures — try again shortly."
    }
    if (brief.data) {
      return `Based on the brief for ${brief.data.date}.`
    }
    return 'Preparing the latest figures…'
  }

  return (
    <Stack spacing={3}>
      <Typography variant="h1" sx={{ fontSize: { xs: '2rem', md: '2.5rem' } }}>
        Reports
      </Typography>
      <Typography variant="body2" color="text.secondary">
        Generate shareable network reports — the Daily Brief plus the network KPIs.
      </Typography>
      <Card variant="outlined">
        <CardContent>
          <Stack spacing={2}>
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              Network report
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {statusText()}
            </Typography>
            <Stack direction="row" spacing={1} sx={{ flexWrap: 'wrap' }}>
              <Button
                variant="contained"
                startIcon={<DownloadOutlined />}
                disabled={!ready}
                onClick={onDownloadMarkdown}
              >
                Download report (Markdown)
              </Button>
              <Button
                variant="outlined"
                startIcon={<DownloadOutlined />}
                disabled={!metrics.data}
                onClick={onDownloadCsv}
              >
                Download KPIs (CSV)
              </Button>
              <Button
                variant="outlined"
                startIcon={<DownloadOutlined />}
                disabled={!ready || pdfBusy}
                onClick={() => void onDownloadPdf()}
              >
                {pdfBusy ? 'Generating PDF…' : 'Download PDF'}
              </Button>
            </Stack>
            {pdfError ? (
              <Alert severity="error">Couldn&apos;t generate the PDF. Try again shortly.</Alert>
            ) : null}
          </Stack>
        </CardContent>
      </Card>

      {brief.data && (
        <Box
          ref={previewRef}
          aria-hidden="true"
          data-testid="report-pdf-preview"
          sx={{
            position: 'absolute',
            left: '-9999px',
            top: 0,
            width: 800,
            p: 4,
            bgcolor: '#ffffff',
          }}
        >
          <Typography variant="h4" gutterBottom>
            Network Report — {brief.data.date}
          </Typography>
          <Typography variant="body1" sx={{ whiteSpace: 'pre-line', mb: 2 }}>
            {brief.data.prose}
          </Typography>
          {brief.data.items.length > 0 && (
            <Stack spacing={1} sx={{ mb: 2 }}>
              <Typography variant="h6">What needs you</Typography>
              {brief.data.items.map((item, idx) => (
                <Typography key={idx} variant="body2">
                  <strong>{item.severity.toUpperCase()} · {item.facility_id}</strong> — {item.headline}
                  {item.explanation ? ` — ${item.explanation}` : ''}
                </Typography>
              ))}
            </Stack>
          )}
          {metrics.data && metrics.data.kpis.length > 0 && (
            <>
              <Typography variant="h6" gutterBottom>
                Network KPIs
              </Typography>
              <Stack spacing={1} sx={{ mb: 2 }}>
                {metrics.data.kpis.map((k) => (
                  <Typography key={k.key} variant="body2">
                    <strong>{k.label}</strong>: {formatKpi(k)}
                  </Typography>
                ))}
              </Stack>
              {chartUrl && (
                <Box
                  component="img"
                  src={chartUrl}
                  alt="Network KPI chart"
                  sx={{ width: '100%', maxWidth: CHART_WIDTH }}
                />
              )}
            </>
          )}
        </Box>
      )}
    </Stack>
  )
}

function formatKpi(k: {
  current: number
  previous: number
  delta_pct: number
  unit: 'count' | 'pesewas' | 'ratio'
  label: string
}): string {
  const delta = `${k.delta_pct >= 0 ? '+' : ''}${k.delta_pct.toFixed(1)}% WoW`
  if (k.unit === 'pesewas') {
    return `${(k.current / 100).toFixed(2)} GHS (${delta})`
  }
  if (k.unit === 'ratio') {
    return `${(k.current * 100).toFixed(1)}% (${delta})`
  }
  return `${Math.round(k.current)} (${delta})`
}

const CHART_WIDTH = 720
