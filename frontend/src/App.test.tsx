import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { App } from './App'

describe('App', () => {
  it('renders the title and seeded status chips', () => {
    render(<App />)
    expect(screen.getByRole('heading', { name: /Gigmann Executive Cockpit/i })).toBeInTheDocument()
    expect(screen.getByText(/Tafo Maternity · CRITICAL/)).toBeInTheDocument()
    expect(screen.getByText(/Adansi · GOOD/)).toBeInTheDocument()
  })

  it('toggles between light and dark mode', () => {
    render(<App />)
    fireEvent.click(screen.getByRole('button', { name: /Toggle dark mode/i }))
    expect(screen.getByRole('button', { name: /Toggle light mode/i })).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /Toggle light mode/i }))
    expect(screen.getByRole('button', { name: /Toggle dark mode/i })).toBeInTheDocument()
  })
})
