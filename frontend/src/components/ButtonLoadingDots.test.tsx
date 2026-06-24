import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { ButtonLoadingDots } from './ButtonLoadingDots'

describe('ButtonLoadingDots', () => {
  it('renders an accessible loading indicator', () => {
    render(<ButtonLoadingDots />)
    expect(screen.getByRole('status', { name: /loading/i })).toBeInTheDocument()
  })
})
