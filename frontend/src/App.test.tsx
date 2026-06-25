import { render, screen, fireEvent } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

vi.mock('./api/useBrief', () => ({
  useBrief: () => ({ data: undefined, isLoading: true, isError: false }),
}))

import { App } from './App'

describe('App', () => {
  it('renders the title and the brief skeleton while loading', () => {
    render(<App />)
    expect(screen.getByRole('heading', { name: /Gigmann Cockpit/i })).toBeInTheDocument()
    expect(screen.getByTestId('brief-skeleton')).toBeInTheDocument()
  })

  it('toggles light/dark mode', () => {
    render(<App />)
    fireEvent.click(screen.getByRole('button', { name: /Dark/i }))
    expect(screen.getByRole('button', { name: /Light/i })).toBeInTheDocument()
  })
})
