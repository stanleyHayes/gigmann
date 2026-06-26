import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { Answer } from '../api/useAsk'
import { AskScreen } from './AskScreen'

type AskState = {
  mutate: ReturnType<typeof vi.fn>
  isPending: boolean
  isError: boolean
  data: Answer | undefined
}

const hoisted = vi.hoisted(() => ({
  result: { mutate: vi.fn(), isPending: false, isError: false, data: undefined } as AskState,
}))

vi.mock('../api/useAsk', () => ({
  useAsk: () => hoisted.result,
}))

describe('AskScreen', () => {
  beforeEach(() => {
    hoisted.result = { mutate: vi.fn(), isPending: false, isError: false, data: undefined }
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

  it('shows a loading indicator while pending', () => {
    hoisted.result = { mutate: vi.fn(), isPending: true, isError: false, data: undefined }
    render(<AskScreen />, { wrapper: MemoryRouter })
    expect(screen.getByLabelText('loading')).toBeInTheDocument()
  })
})
