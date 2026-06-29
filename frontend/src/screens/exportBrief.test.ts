import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { Brief } from '../api/useBrief'
import { answerToText, briefToMarkdown, chartToPng, downloadFile, downloadPdf, networkReportCsv, networkReportMarkdown } from './exportBrief'
import type { NetworkMetrics } from '../api/useMetrics'

vi.mock('html2canvas', () => ({
  default: vi.fn(() =>
    Promise.resolve({
      toDataURL: () => 'data:image/png;base64,pngdata',
    }),
  ),
}))

vi.mock('jspdf', () => ({
  jsPDF: vi.fn(function () {
    return {
      internal: {
        pageSize: {
          getWidth: () => 210,
          getHeight: () => 297,
        },
      },
      addImage: vi.fn(),
      addPage: vi.fn(),
      save: vi.fn(),
    }
  }),
}))

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
  const reportBrief: Brief = { id: 'b', date: '2026-06-09', prose: 'Good morning.', model: 'local', items: [] }
  const metrics: NetworkMetrics = {
    as_of: '2026-06-09',
    kpis: [
      { key: 'revenue', label: 'Revenue', unit: 'pesewas', higher_is_better: true, current: 8_500_000, previous: 8_000_000, delta_pct: 6.25, direction: 'up', series: [] },
      { key: 'denial', label: 'NHIS denial rate', unit: 'ratio', higher_is_better: false, current: 0.19, previous: 0.15, delta_pct: 26.7, direction: 'up', series: [] },
      { key: 'patients', label: 'Patients', unit: 'count', higher_is_better: true, current: 9400, previous: 9300, delta_pct: 1.1, direction: 'up', series: [] },
    ],
  }
  it('includes the brief and KPI lines with formatted values', () => {
    const md = networkReportMarkdown(reportBrief, metrics)
    expect(md).toContain('Good morning.')
    expect(md).toContain('## Network KPIs')
    expect(md).toContain('Revenue')
    expect(md).toContain('19.0%') // ratio formatted as percent
    expect(md).toContain('+6.3% WoW')
  })
  it('omits the KPI section when metrics are absent', () => {
    expect(networkReportMarkdown(reportBrief)).not.toContain('## Network KPIs')
  })
})

describe('networkReportCsv', () => {
  const metrics: NetworkMetrics = {
    as_of: '2026-06-09',
    kpis: [
      {
        key: 'revenue',
        label: 'Revenue',
        unit: 'pesewas',
        higher_is_better: true,
        current: 850000,
        previous: 800000,
        delta_pct: 6.25,
        direction: 'up',
        series: [
          { date: '2026-06-08', value: 800000 },
          { date: '2026-06-09', value: 850000 },
        ],
      },
      {
        key: 'denial_rate',
        label: 'NHIS denial rate',
        unit: 'ratio',
        higher_is_better: false,
        current: 0.195,
        previous: 0.15,
        delta_pct: 30,
        direction: 'up',
        series: [
          { date: '2026-06-08', value: 0.15 },
          { date: '2026-06-09', value: 0.195 },
        ],
      },
      {
        key: 'patients',
        label: 'Patients',
        unit: 'count',
        higher_is_better: true,
        current: 9400,
        previous: 9300,
        delta_pct: 1.1,
        direction: 'up',
        series: [
          { date: '2026-06-08', value: 9300 },
          { date: '2026-06-09', value: 9400 },
        ],
      },
    ],
  }

  it('returns a header and one row per date', () => {
    const csv = networkReportCsv(metrics)
    const lines = csv.split('\n')
    expect(lines[0]).toBe('date,revenue,denial_rate,patients')
    expect(lines).toHaveLength(3)
  })

  it('converts units for spreadsheet use', () => {
    const csv = networkReportCsv(metrics)
    expect(csv).toContain('2026-06-09,8500.00,19.5,9400')
    expect(csv).toContain('2026-06-08,8000.00,15.0,9300')
  })

  it('returns a minimal header when there are no kpis', () => {
    expect(networkReportCsv({ as_of: '2026-06-09', kpis: [] })).toBe('date\n')
  })
})


describe('downloadFile', () => {
  it('creates an anchor and triggers a download', () => {
    const createObjectURL = vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob://x')
    const anchor = document.createElement('a')
    const clickSpy = vi.spyOn(anchor, 'click')
    const createElementSpy = vi.spyOn(document, 'createElement').mockReturnValue(anchor)

    downloadFile('report.md', '# hello')

    expect(createObjectURL).toHaveBeenCalled()
    expect(anchor.download).toBe('report.md')
    expect(anchor.href).toBe('blob://x')
    expect(clickSpy).toHaveBeenCalled()

    createElementSpy.mockRestore()
    createObjectURL.mockRestore()
  })
})

function createFakeCanvas() {
  const ctx = {
    scale: vi.fn(),
    fillRect: vi.fn(),
    fillText: vi.fn(),
    font: '',
    fillStyle: '',
  }
  const canvas = {
    getContext: vi.fn((type: string) => (type === '2d' ? ctx : null)),
    toDataURL: vi.fn(() => 'data:image/png;base64,fakechart'),
    width: 0,
    height: 0,
    style: {} as CSSStyleDeclaration,
  } as unknown as HTMLCanvasElement
  return { canvas, ctx }
}

describe('chartToPng', () => {
  const metrics: NetworkMetrics = {
    as_of: '2026-06-09',
    kpis: [
      {
        key: 'revenue',
        label: 'Revenue',
        unit: 'pesewas',
        higher_is_better: true,
        current: 850000,
        previous: 800000,
        delta_pct: 6.25,
        direction: 'up',
        series: [{ date: '2026-06-09', value: 850000 }],
      },
      {
        key: 'denial_rate',
        label: 'NHIS denial rate',
        unit: 'ratio',
        higher_is_better: false,
        current: 0.195,
        previous: 0.15,
        delta_pct: 30,
        direction: 'up',
        series: [{ date: '2026-06-09', value: 0.15 }],
      },
    ],
  }

  it('returns a PNG data URL', () => {
    const { canvas } = createFakeCanvas()
    const createElementSpy = vi.spyOn(document, 'createElement').mockReturnValue(canvas)

    const url = chartToPng(metrics)

    expect(url).toBe('data:image/png;base64,fakechart')
    expect(canvas.getContext).toHaveBeenCalledWith('2d')
    createElementSpy.mockRestore()
  })

  it('returns an empty string when the canvas context is unavailable', () => {
    const { canvas } = createFakeCanvas()
    canvas.getContext = vi.fn(() => null)
    const createElementSpy = vi.spyOn(document, 'createElement').mockReturnValue(canvas)

    expect(chartToPng(metrics)).toBe('')

    createElementSpy.mockRestore()
  })
})

describe('downloadPdf', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('captures the element and saves a PDF', async () => {
    const { default: html2canvas } = await import('html2canvas')
    const { jsPDF } = await import('jspdf')
    const element = document.createElement('div')

    await downloadPdf('report.pdf', element)

    expect(html2canvas).toHaveBeenCalledWith(element, { scale: 2, backgroundColor: '#ffffff' })
    expect(jsPDF).toHaveBeenCalledWith(expect.objectContaining({ unit: 'mm', format: 'a4' }))
  })
})
