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
