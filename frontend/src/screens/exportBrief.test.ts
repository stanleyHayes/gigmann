import { describe, expect, it } from 'vitest'

import type { Brief } from '../api/useBrief'
import { briefToMarkdown } from './exportBrief'

const brief: Brief = {
  id: 'b1',
  date: '2026-06-26',
  prose: 'Good morning, Sammy.',
  model: 'local',
  items: [
    { severity: 'critical', facility_id: 'tafo-maternity', headline: 'Claims not submitted', explanation: 'demand is flat' },
    { severity: 'watch', facility_id: 'asokwa', headline: 'Stock low' },
  ],
}

describe('briefToMarkdown', () => {
  it('renders the date, prose, and items', () => {
    const md = briefToMarkdown(brief)
    expect(md).toContain('# Daily Brief — 2026-06-26')
    expect(md).toContain('Good morning, Sammy.')
    expect(md).toContain('**CRITICAL · tafo-maternity** — Claims not submitted')
    expect(md).toContain('  demand is flat')
    expect(md).toContain('**WATCH · asokwa** — Stock low')
  })
})
