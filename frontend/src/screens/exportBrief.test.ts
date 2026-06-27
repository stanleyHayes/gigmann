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

import { answerToText, networkReportMarkdown } from './exportBrief'
import type { NetworkMetrics } from '../api/useMetrics'

describe('answerToText', () => {
  it('appends citations as sources', () => {
    expect(answerToText({ text: 'Kasoa denial rate is 19%.', citations: ['kasoa', 'tafo-maternity'] }))
      .toBe('Kasoa denial rate is 19%.\n\nSources: kasoa, tafo-maternity')
  })
  it('omits sources when there are none', () => {
    expect(answerToText({ text: 'All clear.' })).toBe('All clear.')
  })
})

describe('networkReportMarkdown', () => {
  const brief: Brief = { id: 'b', date: '2026-06-09', prose: 'Good morning.', model: 'local', items: [] }
  const metrics: NetworkMetrics = {
    as_of: '2026-06-09',
    kpis: [
      { key: 'revenue', label: 'Revenue', unit: 'pesewas', higher_is_better: true, current: 8_500_000, previous: 8_000_000, delta_pct: 6.25, direction: 'up', series: [] },
      { key: 'denial', label: 'NHIS denial rate', unit: 'ratio', higher_is_better: false, current: 0.19, previous: 0.15, delta_pct: 26.7, direction: 'up', series: [] },
      { key: 'patients', label: 'Patients', unit: 'count', higher_is_better: true, current: 9400, previous: 9300, delta_pct: 1.1, direction: 'up', series: [] },
    ],
  }
  it('includes the brief and KPI lines with formatted values', () => {
    const md = networkReportMarkdown(brief, metrics)
    expect(md).toContain('Good morning.')
    expect(md).toContain('## Network KPIs')
    expect(md).toContain('Revenue')
    expect(md).toContain('19.0%') // ratio formatted as percent
    expect(md).toContain('+6.25% WoW'.replace('6.25', '6.3')) // delta toFixed(1)
  })
  it('omits the KPI section when metrics are absent', () => {
    expect(networkReportMarkdown(brief)).not.toContain('## Network KPIs')
  })
})
