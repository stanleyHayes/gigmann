import { render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { ReportsScreen } from './ReportsScreen'

const hoisted = vi.hoisted(() => ({
  brief: { data: undefined as unknown, isError: false },
  metrics: { data: undefined as unknown },
}))

vi.mock('../api/useBrief', () => ({ useBrief: () => hoisted.brief }))
vi.mock('../api/useMetrics', () => ({ useMetrics: () => hoisted.metrics }))

describe('ReportsScreen', () => {
  beforeEach(() => {
    hoisted.brief = { data: undefined, isError: false }
    hoisted.metrics = { data: undefined }
  })

  it('disables download while the brief is loading', () => {
    render(<ReportsScreen />)
    expect(screen.getByRole('button', { name: /download report/i })).toBeDisabled()
    expect(screen.getByText(/preparing the latest figures/i)).toBeInTheDocument()
  })

  it('enables download when the brief is ready', () => {
    hoisted.brief = {
      data: { id: 'b', date: '2026-06-09', prose: 'x', model: 'local', items: [] },
      isError: false,
    }
    render(<ReportsScreen />)
    expect(screen.getByRole('button', { name: /download report/i })).toBeEnabled()
    expect(screen.getByText(/Based on the brief for 2026-06-09/)).toBeInTheDocument()
  })
})
