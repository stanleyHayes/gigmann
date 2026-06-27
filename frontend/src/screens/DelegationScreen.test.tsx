import { render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { DelegationScreen } from './DelegationScreen'

const hoisted = vi.hoisted(() => ({
  tasks: { data: [] as unknown[], isLoading: false, isError: false },
}))

vi.mock('../api/useTasks', () => ({ useTasks: () => hoisted.tasks }))
vi.mock('../auth/authContext', () => ({ useAuth: () => ({ user: { name: 'Sammy Adjei' } }) }))

describe('DelegationScreen', () => {
  beforeEach(() => {
    hoisted.tasks = { data: [], isLoading: false, isError: false }
  })

  it('shows an empty state when nothing is delegated', () => {
    render(<DelegationScreen />)
    expect(screen.getByText(/nothing delegated/i)).toBeInTheDocument()
  })

  it('groups delegated tasks by assignee and flags stalled ones', () => {
    hoisted.tasks = {
      data: [
        { id: 't1', title: 'Review Kasoa denials', status: 'in_progress', assigned_to: 'Ama Owusu', due_date: '2020-01-01', priority: 'high', source: 'alert', created_at: '2026-06-01' },
        { id: 't2', title: 'My own task', status: 'todo', assigned_to: 'Sammy Adjei', priority: 'low', source: 'manual', created_at: '2026-06-01' },
      ],
      isLoading: false,
      isError: false,
    }
    render(<DelegationScreen />)
    expect(screen.getByText('Ama Owusu')).toBeInTheDocument()
    expect(screen.getByText('Review Kasoa denials')).toBeInTheDocument()
    expect(screen.getByText('Stalled')).toBeInTheDocument()
    expect(screen.queryByText('My own task')).not.toBeInTheDocument()
  })
})
