import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { StatusChip } from './StatusChip'

describe('StatusChip', () => {
  it('prefixes the label when given one', () => {
    render(<StatusChip status="critical" label="Tafo" />)
    expect(screen.getByText(/Tafo · CRITICAL/)).toBeInTheDocument()
  })

  it('shows only the status word without a label', () => {
    render(<StatusChip status="good" />)
    expect(screen.getByText('GOOD')).toBeInTheDocument()
  })
})
