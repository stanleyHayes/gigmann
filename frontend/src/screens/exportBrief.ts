import type { Brief } from '../api/useBrief'

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
