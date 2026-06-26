import { render, screen, fireEvent } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { Task } from '../api/useTasks'
import { MyDayScreen } from './MyDayScreen'

const hoisted = vi.hoisted(() => ({
  tasks: { data: undefined as Task[] | undefined, isLoading: true, isError: false },
  mutate: vi.fn(),
}))

vi.mock('../api/useTasks', () => ({
  useTasks: () => hoisted.tasks,
  useUpdateTaskStatus: () => ({ mutate: hoisted.mutate, isPending: false }),
}))

const todo: Task = {
  id: 'task-tafo', title: 'Message Tafo manager', detail: 'Claims not submitted',
  facility_id: 'tafo-maternity', priority: 'high', status: 'todo', source: 'brief',
  due_date: '2026-06-26', created_at: '2026-06-26T00:00:00Z',
}
const done: Task = {
  id: 'task-deck', title: 'Finalise board deck', priority: 'medium', status: 'done',
  source: 'manual', created_at: '2026-06-26T00:00:00Z',
}

describe('MyDayScreen', () => {
  beforeEach(() => {
    hoisted.tasks = { data: undefined, isLoading: true, isError: false }
    hoisted.mutate.mockReset()
  })

  it('shows a skeleton while loading', () => {
    render(<MyDayScreen />)
    expect(screen.getByTestId('myday-skeleton')).toBeInTheDocument()
  })

  it('shows an empty state', () => {
    hoisted.tasks = { data: [], isLoading: false, isError: false }
    render(<MyDayScreen />)
    expect(screen.getByText(/nothing on your list/i)).toBeInTheDocument()
  })

  it('lists tasks and completes one via the checkbox', () => {
    hoisted.tasks = { data: [todo], isLoading: false, isError: false }
    render(<MyDayScreen />)
    expect(screen.getByText('Message Tafo manager')).toBeInTheDocument()

    const checkbox = screen.getByRole('checkbox', { name: /mark .* done/i })
    expect(checkbox).not.toBeChecked()
    fireEvent.click(checkbox)
    expect(hoisted.mutate).toHaveBeenCalledTimes(1)
    expect(hoisted.mutate.mock.calls[0][0]).toEqual({ id: 'task-tafo', status: 'done' })
  })

  it('shows a completed task as checked', () => {
    hoisted.tasks = { data: [done], isLoading: false, isError: false }
    render(<MyDayScreen />)
    expect(screen.getByRole('checkbox')).toBeChecked()
  })
})
