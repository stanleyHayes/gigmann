import type { Brief } from '../api/useBrief'
import type { Kpi, MetricPoint, NetworkMetrics } from '../api/useMetrics'
import { fmt } from '../i18n/locale'

/** briefToMarkdown renders a Daily Brief as shareable Markdown. */
export function briefToMarkdown(brief: Brief): string {
  const lines: string[] = [`# Daily Brief — ${brief.date}`, '', brief.prose, '']
  if (brief.items.length > 0) {
    lines.push('## What needs you', '')
    for (const item of brief.items) {
      lines.push(`- **${item.severity.toUpperCase()} · ${item.facility_id}** — ${item.headline}`)
      if (item.explanation) {
        lines.push(`  ${item.explanation}`)
      }
    }
    lines.push('')
  }
  return lines.join('\n')
}

/** answerToText renders a grounded Ask answer (with its citations) as shareable text. */
export function answerToText(answer: { text: string; citations?: string[] }): string {
  const sources = answer.citations && answer.citations.length > 0 ? `\n\nSources: ${answer.citations.join(', ')}` : ''
  return `${answer.text}${sources}`
}

function formatKpiCurrent(k: Kpi): string {
  if (k.unit === 'pesewas') return fmt.cedis(k.current / 100)
  if (k.unit === 'ratio') return `${(k.current * 100).toFixed(1)}%`
  return fmt.number(Math.round(k.current))
}

function rawKpiValue(value: number, unit: Kpi['unit']): string {
  if (unit === 'pesewas') return (value / 100).toFixed(2)
  if (unit === 'ratio') return (value * 100).toFixed(1)
  return String(Math.round(value))
}

/** networkReportMarkdown combines the Daily Brief and the network KPIs into one report. */
export function networkReportMarkdown(brief: Brief, metrics?: NetworkMetrics): string {
  const lines = [briefToMarkdown(brief)]
  if (metrics && metrics.kpis.length > 0) {
    lines.push('', '## Network KPIs', '')
    for (const k of metrics.kpis) {
      const delta = `${k.delta_pct >= 0 ? '+' : ''}${k.delta_pct.toFixed(1)}% WoW`
      lines.push(`- **${k.label}**: ${formatKpiCurrent(k)} (${delta})`)
    }
  }
  return lines.join('\n')
}

/**
 * networkReportCsv returns the network KPI series as a CSV string (one row per
 * date, one column per KPI) so it can be opened in a spreadsheet.
 */
export function networkReportCsv(metrics: NetworkMetrics): string {
  const kpis = metrics.kpis
  if (kpis.length === 0 || kpis[0].series.length === 0) {
    return 'date\n'
  }

  const dates = kpis[0].series.map((p: MetricPoint) => p.date)
  const headers = ['date', ...kpis.map((k) => k.key)]
  const rows = dates.map((date, index) => [
    date,
    ...kpis.map((k) => rawKpiValue(k.series[index]?.value ?? 0, k.unit)),
  ])
  return [headers.join(','), ...rows.map((r) => r.join(','))].join('\n')
}

/** downloadFile triggers a client-side download of text content. */
export function downloadFile(filename: string, content: string, mime = 'text/markdown'): void {
  const url = URL.createObjectURL(new Blob([content], { type: mime }))
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

const CHART_WIDTH = 720
const CHART_ROW = 120
const CHART_PADDING = 24

/** chartToPng draws the network KPIs as a simple bar chart and returns a PNG data URL. */
export function chartToPng(metrics: NetworkMetrics): string {
  if (typeof document === 'undefined') return ''
  const canvas = document.createElement('canvas')
  const dpr = typeof window !== 'undefined' ? window.devicePixelRatio || 1 : 1
  const height = CHART_PADDING * 2 + metrics.kpis.length * CHART_ROW
  canvas.width = Math.round(CHART_WIDTH * dpr)
  canvas.height = Math.round(height * dpr)
  canvas.style.width = `${CHART_WIDTH}px`
  canvas.style.height = `${height}px`

  const ctx = canvas.getContext('2d')
  if (!ctx) return ''

  ctx.scale(dpr, dpr)
  ctx.fillStyle = '#ffffff'
  ctx.fillRect(0, 0, CHART_WIDTH, height)

  ctx.fillStyle = '#111827'
  ctx.font = 'bold 18px sans-serif'
  ctx.fillText('Network KPIs', CHART_PADDING, CHART_PADDING - 4)

  metrics.kpis.forEach((k, i) => {
    const y = CHART_PADDING + i * CHART_ROW
    ctx.fillStyle = '#111827'
    ctx.font = 'bold 14px sans-serif'
    ctx.fillText(k.label, CHART_PADDING, y + 18)

    const maxValue = Math.max(k.current, k.previous, ...k.series.map((p) => p.value), 1)
    const barY = y + 34
    const barMaxW = CHART_WIDTH - CHART_PADDING * 2 - 160
    const curW = (k.current / maxValue) * barMaxW
    const prevW = (k.previous / maxValue) * barMaxW

    ctx.fillStyle = '#1976d2'
    ctx.fillRect(CHART_PADDING, barY, curW, 22)
    ctx.fillStyle = '#9e9e9e'
    ctx.fillRect(CHART_PADDING, barY + 28, prevW, 22)

    ctx.fillStyle = '#111827'
    ctx.font = '12px sans-serif'
    ctx.fillText(`${formatKpiCurrent(k)} current`, CHART_PADDING + curW + 8, barY + 16)
    ctx.fillText(`${rawKpiValue(k.previous, k.unit)} previous`, CHART_PADDING + prevW + 8, barY + 44)
  })

  try {
    return canvas.toDataURL('image/png')
  } catch {
    return ''
  }
}

/** downloadPdf renders the supplied element to a PNG via html2canvas and saves it as a PDF. */
export async function downloadPdf(filename: string, element: HTMLElement): Promise<void> {
  const [{ default: html2canvas }, { jsPDF }] = await Promise.all([
    import('html2canvas'),
    import('jspdf'),
  ])

  const canvas = await html2canvas(element, { scale: 2, backgroundColor: '#ffffff' })
  const imgData = canvas.toDataURL('image/png')

  const pdf = new jsPDF({ unit: 'mm', format: 'a4' })
  const pageWidth = pdf.internal.pageSize.getWidth()
  const pageHeight = pdf.internal.pageSize.getHeight()
  const margin = 10
  const imgWidth = pageWidth - margin * 2
  const imgHeight = (canvas.height * imgWidth) / canvas.width

  let heightLeft = imgHeight
  let position = margin
  pdf.addImage(imgData, 'PNG', margin, position, imgWidth, imgHeight)
  heightLeft -= pageHeight - margin * 2

  while (heightLeft > 0) {
    pdf.addPage()
    position = heightLeft - imgHeight + margin
    pdf.addImage(imgData, 'PNG', margin, position, imgWidth, imgHeight)
    heightLeft -= pageHeight - margin * 2
  }

  pdf.save(filename)
}
