import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { Answer } from '../api/useAsk'
import type { Facility } from '../api/useFacilities'
import { AskScreen } from './AskScreen'

type AskState = {
  mutate: ReturnType<typeof vi.fn>
  isPending: boolean
  isError: boolean
  data: Answer | undefined
}

const hoisted = vi.hoisted(() => ({
  result: { mutate: vi.fn(), isPending: false, isError: false, data: undefined } as AskState,
  draft: { mutate: vi.fn(), isPending: false, isError: false, data: undefined as { draft: string } | undefined },
  facilities: [] as Facility[],
}))

vi.mock('../api/useAsk', () => ({
  useAsk: () => hoisted.result,
}))
vi.mock('../api/useDrafts', () => ({
  useCreateDraft: () => hoisted.draft,
}))
vi.mock('../api/useFacilities', () => ({
  useFacilities: () => ({ data: hoisted.facilities }),
}))

describe('AskScreen', () => {
  beforeEach(() => {
    hoisted.result = { mutate: vi.fn(), isPending: false, isError: false, data: undefined }
    hoisted.draft = { mutate: vi.fn(), isPending: false, isError: false, data: undefined }
    hoisted.facilities = []
  })

  it('asks the typed question', () => {
    render(<AskScreen />, { wrapper: MemoryRouter })
    fireEvent.change(screen.getByLabelText('Question'), { target: { value: 'Why is Kasoa bad?' } })
    fireEvent.click(screen.getByRole('button', { name: /^ask$/i }))
    expect(hoisted.result.mutate).toHaveBeenCalledWith('Why is Kasoa bad?')
  })

  it('asks a suggested question on click', () => {
    render(<AskScreen />, { wrapper: MemoryRouter })
    fireEvent.click(screen.getByText('What is driving the NHIS denials?'))
    expect(hoisted.result.mutate).toHaveBeenCalledWith('What is driving the NHIS denials?')
  })

  it('renders the grounded answer with citations', () => {
    hoisted.result = {
      mutate: vi.fn(),
      isPending: false,
      isError: false,
      data: { text: 'Kasoa is worst — 20% denial rate.', citations: ['kasoa'] },
    }
    render(<AskScreen />, { wrapper: MemoryRouter })
    expect(screen.getByText(/Kasoa is worst/)).toBeInTheDocument()
    expect(screen.getByText('kasoa')).toBeInTheDocument()
  })

  it('generates a controlled draft', () => {
    hoisted.facilities = [
      { id: 'kasoa', name: 'Kasoa Diagnostics', region: 'Central', town: 'Kasoa', beds: 30, status: 'watch' },
    ]
    render(<AskScreen />, { wrapper: MemoryRouter })
    fireEvent.mouseDown(screen.getByLabelText(/facility/i))
    fireEvent.click(screen.getByRole('option', { name: /Kasoa Diagnostics/i }))
    fireEvent.change(screen.getByLabelText(/instruction/i), { target: { value: 'Summarise the denial trend.' } })
    fireEvent.click(screen.getByRole('button', { name: /generate draft/i }))
    expect(hoisted.draft.mutate).toHaveBeenCalledWith({
      kind: 'message',
      facility_id: 'kasoa',
      instruction: 'Summarise the denial trend.',
    })
  })

  it('shows a loading indicator while pending', () => {
    hoisted.result = { mutate: vi.fn(), isPending: true, isError: false, data: undefined }
    render(<AskScreen />, { wrapper: MemoryRouter })
    expect(screen.getByLabelText('loading')).toBeInTheDocument()
  })

  it('copies a generated draft to the clipboard', () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true })
    hoisted.draft = { mutate: vi.fn(), isPending: false, isError: false, data: { draft: 'Dear manager, reorder RDT kits.' } }
    render(<AskScreen />, { wrapper: MemoryRouter })
    fireEvent.click(screen.getByRole('button', { name: /copy draft/i }))
    expect(writeText).toHaveBeenCalledWith('Dear manager, reorder RDT kits.')
  })
})
