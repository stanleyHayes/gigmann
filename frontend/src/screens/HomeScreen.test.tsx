import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { Brief } from '../api/useBrief'
import { HomeScreen } from './HomeScreen'

const hoisted = vi.hoisted(() => ({
  result: { data: undefined as Brief | undefined, isLoading: true, isError: false },
}))

vi.mock('../api/useBrief', () => ({
  useBrief: () => hoisted.result,
}))

const brief: Brief = {
  id: 'b1', date: '2026-06-26', prose: 'Good morning, Sammy.', model: 'local',
  items: [{ severity: 'critical', facility_id: 'tafo-maternity', headline: 'Claims not submitted' }],
}

describe('HomeScreen', () => {
  beforeEach(() => {
    hoisted.result = { data: undefined, isLoading: true, isError: false }
  })

  it('hides export actions until the brief loads', () => {
    render(<HomeScreen />, { wrapper: MemoryRouter })
    expect(screen.queryByRole('button', { name: /copy/i })).not.toBeInTheDocument()
  })

  it('copies the brief as markdown', () => {
    const writeText = vi.fn()
    vi.stubGlobal('navigator', { clipboard: { writeText } })
    hoisted.result = { data: brief, isLoading: false, isError: false }

    render(<HomeScreen />, { wrapper: MemoryRouter })
    fireEvent.click(screen.getByRole('button', { name: /copy/i }))
    expect(writeText).toHaveBeenCalledWith(expect.stringContaining('# Daily Brief — 2026-06-26'))
    vi.unstubAllGlobals()
  })
})
