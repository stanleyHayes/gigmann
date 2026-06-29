import { fireEvent, render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { chartToPng, downloadPdf } from './exportBrief'
import { ReportsScreen } from './ReportsScreen'

const hoisted = vi.hoisted(() => ({
  brief: { data: undefined as unknown, isError: false },
  metrics: { data: undefined as unknown },
}))

vi.mock('../api/useBrief', () => ({ useBrief: () => hoisted.brief }))
vi.mock('../api/useMetrics', () => ({ useMetrics: () => hoisted.metrics }))

vi.mock('./exportBrief', async () => {
  const actual = await vi.importActual<typeof import('./exportBrief')>('./exportBrief')
  return {
    ...actual,
    chartToPng: vi.fn(() => 'data:image/png;base64,chart'),
    downloadPdf: vi.fn(() => Promise.resolve()),
  }
})

describe('ReportsScreen', () => {
  beforeEach(() => {
    hoisted.brief = { data: undefined, isError: false }
    hoisted.metrics = { data: undefined }
    vi.mocked(downloadPdf).mockClear()
    vi.mocked(chartToPng).mockClear()
  })

  it('disables downloads while the brief is loading', () => {
    render(<ReportsScreen />)
    expect(screen.getByRole('button', { name: /download report/i })).toBeDisabled()
    expect(screen.getByRole('button', { name: /download kpis/i })).toBeDisabled()
    expect(screen.getByRole('button', { name: /download pdf/i })).toBeDisabled()
    expect(screen.getByText(/preparing the latest figures/i)).toBeInTheDocument()
  })

  it('enables downloads when data is ready', () => {
    hoisted.brief = {
      data: {
        id: 'b',
        date: '2026-06-09',
        prose: 'x',
        model: 'local',
        items: [{ severity: 'critical', facility_id: 'tafo-maternity', headline: 'Claims not submitted', explanation: 'demand is flat' }],
      },
      isError: false,
    }
    hoisted.metrics = {
      data: {
        as_of: '2026-06-09',
        kpis: [
          {
            key: 'revenue',
            label: 'Revenue',
            unit: 'pesewas',
            higher_is_better: true,
            current: 100000,
            previous: 90000,
            delta_pct: 11.1,
            direction: 'up',
            series: [{ date: '2026-06-09', value: 100000 }],
          },
          {
            key: 'denial_rate',
            label: 'NHIS denial rate',
            unit: 'ratio',
            higher_is_better: false,
            current: 0.19,
            previous: 0.15,
            delta_pct: 26.7,
            direction: 'up',
            series: [{ date: '2026-06-09', value: 0.19 }],
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
            series: [{ date: '2026-06-09', value: 9400 }],
          },
        ],
      },
    }
    render(<ReportsScreen />)
    expect(screen.getByRole('button', { name: /download report/i })).toBeEnabled()
    expect(screen.getByRole('button', { name: /download kpis/i })).toBeEnabled()
    expect(screen.getByRole('button', { name: /download pdf/i })).toBeEnabled()
    expect(screen.getByText(/Based on the brief for 2026-06-09/)).toBeInTheDocument()
    expect(screen.getByTestId('report-pdf-preview')).toBeInTheDocument()
    expect(screen.getByText(/demand is flat/)).toBeInTheDocument()
    expect(chartToPng).toHaveBeenCalledWith(hoisted.metrics.data)
  })

  it('downloads a PDF from the hidden preview when the PDF button is clicked', async () => {
    hoisted.brief = {
      data: {
        id: 'b',
        date: '2026-06-09',
        prose: 'x',
        model: 'local',
        items: [{ severity: 'critical', facility_id: 'tafo-maternity', headline: 'Claims not submitted', explanation: 'demand is flat' }],
      },
      isError: false,
    }
    hoisted.metrics = {
      data: {
        as_of: '2026-06-09',
        kpis: [
          {
            key: 'revenue',
            label: 'Revenue',
            unit: 'pesewas',
            higher_is_better: true,
            current: 100000,
            previous: 90000,
            delta_pct: 11.1,
            direction: 'up',
            series: [{ date: '2026-06-09', value: 100000 }],
          },
          {
            key: 'denial_rate',
            label: 'NHIS denial rate',
            unit: 'ratio',
            higher_is_better: false,
            current: 0.19,
            previous: 0.15,
            delta_pct: 26.7,
            direction: 'up',
            series: [{ date: '2026-06-09', value: 0.19 }],
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
            series: [{ date: '2026-06-09', value: 9400 }],
          },
        ],
      },
    }
    render(<ReportsScreen />)
    const pdfButton = screen.getByRole('button', { name: /download pdf/i })
    fireEvent.click(pdfButton)
    expect(downloadPdf).toHaveBeenCalledTimes(1)
    const preview = screen.getByTestId('report-pdf-preview')
    expect(downloadPdf).toHaveBeenCalledWith('network-report-2026-06-09.pdf', preview)
  })
})
