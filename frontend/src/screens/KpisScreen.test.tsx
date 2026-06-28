import { render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { Kpi, NetworkMetrics } from '../api/useMetrics'
import { KpiCard, KpisScreen } from './KpisScreen'

vi.mock('@mui/x-charts/LineChart', () => ({
  LineChart: () => <div data-testid="line-chart" />,
}))

const hoisted = vi.hoisted(() => ({
  result: { data: undefined as NetworkMetrics | undefined, isLoading: true, isError: false },
}))

vi.mock('../api/useMetrics', () => ({
  useMetrics: () => hoisted.result,
}))

const revenue: Kpi = {
  key: 'revenue', label: 'Network revenue', unit: 'pesewas', higher_is_better: true,
  current: 114337612, previous: 100000000, delta_pct: 0.14, direction: 'up',
  series: [{ date: '2026-06-12', value: 8000000 }, { date: '2026-06-13', value: 8200000 }],
}
const denial: Kpi = {
  key: 'denial_rate', label: 'NHIS denial rate', unit: 'ratio', higher_is_better: false,
  current: 0.07, previous: 0.05, delta_pct: 0.35, direction: 'up',
  series: [{ date: '2026-06-12', value: 0.05 }, { date: '2026-06-13', value: 0.07 }],
}
const patients: Kpi = {
  key: 'patients', label: 'Patients seen', unit: 'count', higher_is_better: true,
  current: 14724, previous: 14724, delta_pct: 0, direction: 'flat',
  series: [{ date: '2026-06-12', value: 1000 }, { date: '2026-06-13', value: 1050 }],
}
const occupancy: Kpi = {
  key: 'occupancy', label: 'Bed occupancy', unit: 'ratio', higher_is_better: true,
  current: 0.7, previous: 0.72, delta_pct: -0.03, direction: 'down',
  series: [{ date: '2026-06-12', value: 0.72 }, { date: '2026-06-13', value: 0.7 }],
}

describe('KpiCard', () => {
  it('formats a money KPI and flags an improvement', () => {
    render(<KpiCard kpi={revenue} reduceMotion={false} />)
    // Pesewas are preserved (no silent truncation of the .12).
    expect(screen.getByText(/1,143,376\.12/)).toBeInTheDocument()
    expect(screen.getByText(/14%/)).toBeInTheDocument()
    expect(screen.getByLabelText('improved')).toBeInTheDocument()
    expect(screen.getByTestId('line-chart')).toBeInTheDocument()
  })

  it('formats a ratio KPI and flags a worsening', () => {
    render(<KpiCard kpi={denial} reduceMotion />)
    expect(screen.getByText('7.0%')).toBeInTheDocument()
    expect(screen.getByLabelText('worsened')).toBeInTheDocument()
  })

  it('treats a flat KPI as unchanged', () => {
    render(<KpiCard kpi={patients} reduceMotion={false} />)
    expect(screen.getByText('14,724')).toBeInTheDocument()
    expect(screen.getByLabelText('unchanged')).toBeInTheDocument()
  })
})

describe('KpisScreen', () => {
  beforeEach(() => {
    hoisted.result = { data: undefined, isLoading: true, isError: false }
  })

  it('shows a skeleton while loading', () => {
    render(<KpisScreen />)
    expect(screen.getByTestId('kpis-skeleton')).toBeInTheDocument()
  })

  it('shows an error state', () => {
    hoisted.result = { data: undefined, isLoading: false, isError: true }
    render(<KpisScreen />)
    expect(screen.getByText(/couldn.t load the kpis/i)).toBeInTheDocument()
  })

  it('renders a card per KPI when loaded', () => {
    hoisted.result = {
      data: { as_of: '2026-06-25', kpis: [revenue, denial, patients, occupancy] },
      isLoading: false,
      isError: false,
    }
    render(<KpisScreen />)
    expect(screen.getByRole('heading', { name: /Executive KPIs/i })).toBeInTheDocument()
    expect(screen.getByText('Network revenue')).toBeInTheDocument()
    expect(screen.getByText('Bed occupancy')).toBeInTheDocument()
    expect(screen.getAllByTestId('line-chart')).toHaveLength(4)
  })
})
