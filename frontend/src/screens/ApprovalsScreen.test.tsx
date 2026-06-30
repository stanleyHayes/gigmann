import { render, screen, fireEvent } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { Approval } from '../api/useApprovals'
import { ApprovalsScreen } from './ApprovalsScreen'

const hoisted = vi.hoisted(() => ({
  approvals: { data: undefined as Approval[] | undefined, isLoading: true, isError: false },
  mutate: vi.fn(),
}))

vi.mock('../api/useApprovals', () => ({
  useApprovals: () => hoisted.approvals,
  useDecideApproval: () => ({ mutate: hoisted.mutate, isPending: false }),
}))

const pending: Approval = {
  id: 'ap1', type: 'capital', facility_id: 'assin-fosu', amount_pesewas: 8500000,
  title: 'Ultrasound machine', context: 'Replaces ageing unit', requested_by: 'Dr. Mensah',
  status: 'pending', created_at: '2026-06-26T00:00:00Z',
}
const approved: Approval = {
  id: 'ap2', type: 'reorder', facility_id: 'nima', amount_pesewas: 1400000,
  title: 'Generator servicing', requested_by: 'Mohammed', status: 'approved',
  decision_note: 'Approved for Q3', created_at: '2026-06-26T00:00:00Z',
}

describe('ApprovalsScreen', () => {
  beforeEach(() => {
    hoisted.approvals = { data: undefined, isLoading: true, isError: false }
    hoisted.mutate.mockReset()
  })

  it('shows a skeleton while loading', () => {
    render(<ApprovalsScreen />)
    expect(screen.getByTestId('approvals-skeleton')).toBeInTheDocument()
  })

  it('shows an empty state', () => {
    hoisted.approvals = { data: [], isLoading: false, isError: false }
    render(<ApprovalsScreen />)
    expect(screen.getByText(/queue is clear/i)).toBeInTheDocument()
    expect(screen.getByText(/reorder approvals/i)).toBeInTheDocument()
  })

  it('requires explicit confirmation before deciding', () => {
    hoisted.approvals = { data: [pending], isLoading: false, isError: false }
    render(<ApprovalsScreen />)

    expect(screen.getByText('Ultrasound machine')).toBeInTheDocument()
    // clicking Approve must NOT decide immediately — it opens a confirm dialog
    fireEvent.click(screen.getByRole('button', { name: 'Approve' }))
    expect(hoisted.mutate).not.toHaveBeenCalled()
    expect(screen.getByRole('dialog')).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /confirm approve/i }))
    expect(hoisted.mutate).toHaveBeenCalledTimes(1)
    expect(hoisted.mutate.mock.calls[0][0]).toEqual({ id: 'ap1', decision: 'approve', note: undefined })
  })

  it('shows the decision on a settled approval and hides the controls', () => {
    hoisted.approvals = { data: [approved], isLoading: false, isError: false }
    render(<ApprovalsScreen />)
    expect(screen.getByText('approved')).toBeInTheDocument()
    expect(screen.getByText(/note: approved for q3/i)).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Approve' })).not.toBeInTheDocument()
  })

  it('renders the amount without truncating fractional cedis', () => {
    // 8_500_050 pesewas = GH₵85,000.50. The old Math.round(pesewas/100) would
    // have rounded to 85,001 and dropped the pesewa precision entirely.
    const fractional: Approval = { ...pending, amount_pesewas: 8_500_050 }
    hoisted.approvals = { data: [fractional], isLoading: false, isError: false }
    render(<ApprovalsScreen />)
    expect(screen.getByText(/85,000\.50/)).toBeInTheDocument()
    expect(screen.queryByText(/85,001/)).not.toBeInTheDocument()
  })
})
