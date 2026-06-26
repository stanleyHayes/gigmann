import { render, screen, fireEvent } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import type { Brief } from '../api/useBrief'
import { DailyBrief } from './DailyBrief'

const brief: Brief = {
  id: 'b1',
  date: '2026-06-09',
  prose: 'Good morning, Sammy.',
  model: 'local-deterministic',
  items: [
    {
      severity: 'critical',
      facility_id: 'tafo-maternity',
      headline: 'Tafo needs you first',
      explanation: 'Claims recorded but not submitted',
      suggested_actions: ['Why?', 'Message the manager'],
    },
    { severity: 'watch', facility_id: 'asokwa', headline: 'Stock running low' },
  ],
}

describe('DailyBrief', () => {
  it('shows a skeleton while loading', () => {
    render(<DailyBrief isLoading isError={false} />)
    expect(screen.getByTestId('brief-skeleton')).toBeInTheDocument()
  })

  it('shows an error state', () => {
    render(<DailyBrief isLoading={false} isError />)
    expect(screen.getByText(/couldn.t load the brief/i)).toBeInTheDocument()
  })

  it('renders the prose and prioritised items', () => {
    render(<DailyBrief brief={brief} isLoading={false} isError={false} />)
    expect(screen.getByText('Good morning, Sammy.')).toBeInTheDocument()
    expect(screen.getByText('Tafo needs you first')).toBeInTheDocument()
    expect(screen.getByText('Claims recorded but not submitted')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Why\?/ })).toBeInTheDocument()
    expect(screen.getByText('Stock running low')).toBeInTheDocument()
  })
})

describe('DailyBrief actions', () => {
  it('invokes onAction when a suggested action is clicked', () => {
    const onAction = vi.fn()
    render(<DailyBrief brief={brief} isLoading={false} isError={false} onAction={onAction} />)
    fireEvent.click(screen.getByRole('button', { name: /Why\? for tafo-maternity/i }))
    expect(onAction).toHaveBeenCalledWith('Why?', 'tafo-maternity')
  })
})
