import { fireEvent, render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { AlertItem } from '../api/useAlerts'

const h = vi.hoisted(() => ({
  alerts: { data: { alerts: [] as AlertItem[] } as { alerts: AlertItem[] } | undefined, isLoading: false, isError: false },
  updateMutate: vi.fn(),
  createMutate: vi.fn(),
  createReset: vi.fn(),
}))

vi.mock('../api/useAlerts', () => ({
  useAlerts: () => h.alerts,
  useUpdateAlertStatus: () => ({ mutate: h.updateMutate, isPending: false }),
}))
vi.mock('../api/useTasks', () => ({
  useCreateTask: () => ({ mutate: h.createMutate, isPending: false, isError: false, reset: h.createReset }),
}))

import { AttentionFeed } from './AttentionFeed'

const alert: AlertItem = {
  id: 'al1',
  facility_id: 'kasoa',
  type: 'revenue_drop',
  severity: 'critical',
  status: 'open',
  title: 'Kasoa revenue down 22%',
  detail: 'Demand flat',
  created_at: '2026-06-30T00:00:00Z',
} as AlertItem

describe('AttentionFeed', () => {
  beforeEach(() => {
    h.alerts = { data: { alerts: [] }, isLoading: false, isError: false }
    h.updateMutate.mockReset()
    h.createMutate.mockReset()
    h.createReset.mockReset()
  })

  it('shows a skeleton while loading', () => {
    h.alerts = { data: undefined, isLoading: true, isError: false }
    render(<AttentionFeed />)
    expect(screen.getByTestId('attention-feed-skeleton')).toBeInTheDocument()
  })

  it('shows an error state', () => {
    h.alerts = { data: undefined, isLoading: false, isError: true }
    render(<AttentionFeed />)
    expect(screen.getByText(/couldn.t load the attention feed/i)).toBeInTheDocument()
  })

  it('shows an empty state when there are no open alerts', () => {
    render(<AttentionFeed />)
    expect(screen.getByText(/no open alerts/i)).toBeInTheDocument()
  })

  it('renders an alert and wires the task/resolve/dismiss actions', () => {
    h.alerts = { data: { alerts: [alert] }, isLoading: false, isError: false }
    render(<AttentionFeed />)
    expect(screen.getByText('Kasoa revenue down 22%')).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /turn into task/i }))
    expect(h.createMutate).toHaveBeenCalledWith(
      expect.objectContaining({ title: 'Kasoa revenue down 22%', facility_id: 'kasoa', priority: 'high', source: 'alert' }),
      expect.anything(),
    )

    fireEvent.click(screen.getByRole('button', { name: /resolve/i }))
    expect(h.updateMutate).toHaveBeenCalledWith({ id: 'al1', status: 'resolved' })

    fireEvent.click(screen.getByRole('button', { name: /dismiss/i }))
    expect(h.updateMutate).toHaveBeenCalledWith({ id: 'al1', status: 'dismissed' })
  })
})
