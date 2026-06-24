import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { StatusChip } from './StatusChip'

describe('StatusChip', () => {
  it('renders the facility label with an uppercase status (colour + text, a11y)', () => {
    render(<StatusChip status="critical" label="Tafo Maternity" />)
    expect(screen.getByText(/Tafo Maternity · CRITICAL/)).toBeInTheDocument()
  })

  it('renders the good status', () => {
    render(<StatusChip status="good" label="Adansi" />)
    expect(screen.getByText(/Adansi · GOOD/)).toBeInTheDocument()
  })
})
