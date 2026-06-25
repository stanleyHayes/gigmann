import { render, screen, fireEvent } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

vi.mock('./api/useBrief', () => ({
  useBrief: () => ({ data: undefined, isLoading: true, isError: false }),
}))

import { App } from './App'

describe('App', () => {
  it('renders the cockpit shell with the brief on the index route', () => {
    render(<App />)
    expect(screen.getByText('Ahenfie')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: /The Brief/i })).toBeInTheDocument()
    expect(screen.getByTestId('brief-skeleton')).toBeInTheDocument()
  })

  it('exposes primary navigation', () => {
    render(<App />)
    expect(screen.getByRole('link', { name: /Today/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Network/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Approvals/i })).toBeInTheDocument()
  })

  it('toggles the colour mode', () => {
    render(<App />)
    const toggle = screen.getByRole('button', { name: /Switch to dark mode/i })
    fireEvent.click(toggle)
    expect(screen.getByRole('button', { name: /Switch to light mode/i })).toBeInTheDocument()
  })
})
