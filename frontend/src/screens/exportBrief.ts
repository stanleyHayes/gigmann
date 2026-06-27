import type { Brief } from '../api/useBrief'
import type { Kpi, NetworkMetrics } from '../api/useMetrics'
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

function kpiValue(k: Kpi): string {
  if (k.unit === 'pesewas') return fmt.cedis(k.current / 100)
  if (k.unit === 'ratio') return `${(k.current * 100).toFixed(1)}%`
  return fmt.number(Math.round(k.current))
}

/** networkReportMarkdown combines the Daily Brief and the network KPIs into one report. */
export function networkReportMarkdown(brief: Brief, metrics?: NetworkMetrics): string {
  const lines = [briefToMarkdown(brief)]
  if (metrics && metrics.kpis.length > 0) {
    lines.push('', '## Network KPIs', '')
    for (const k of metrics.kpis) {
      const delta = `${k.delta_pct >= 0 ? '+' : ''}${k.delta_pct.toFixed(1)}% WoW`
      lines.push(`- **${k.label}**: ${kpiValue(k)} (${delta})`)
    }
  }
  return lines.join('\n')
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
