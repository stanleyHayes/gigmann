import { render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { Facility } from '../api/useFacilities'
import { NetworkOverview, NetworkScreen } from './NetworkScreen'

const hoisted = vi.hoisted(() => ({
  result: { data: undefined as Facility[] | undefined, isLoading: true, isError: false },
}))

vi.mock('../api/useFacilities', () => ({
  useFacilities: () => hoisted.result,
}))

const facilities: Facility[] = [
  { id: 'a', name: 'Asokwa Clinic', region: 'Ashanti', town: 'Asokwa', beds: 20, status: 'good' },
  { id: 'b', name: 'Tafo Maternity', region: 'Ashanti', town: 'Tafo', beds: 40, status: 'critical' },
  { id: 'c', name: 'Kasoa Diagnostics', region: 'Central', town: 'Kasoa', beds: 30, status: 'watch' },
]

describe('NetworkOverview', () => {
  it('summarises the network and sorts facilities worst-first', () => {
    render(<NetworkOverview facilities={facilities} />)
    expect(screen.getByText(/3 facilities/i)).toBeInTheDocument()
    expect(screen.getByText(/1 critical/i)).toBeInTheDocument()

    const names = screen.getAllByRole('heading', { level: 6 }).map((h) => h.textContent)
    expect(names[0]).toMatch(/Tafo/) // critical first
    expect(names[names.length - 1]).toMatch(/Asokwa/) // healthy last
  })

  it('handles an empty network', () => {
    render(<NetworkOverview facilities={[]} />)
    expect(screen.getByText(/0 facilities/i)).toBeInTheDocument()
  })
})

describe('NetworkScreen', () => {
  beforeEach(() => {
    hoisted.result = { data: undefined, isLoading: true, isError: false }
  })

  it('shows a skeleton while loading', () => {
    render(<NetworkScreen />)
    expect(screen.getByTestId('network-skeleton')).toBeInTheDocument()
  })

  it('shows an error state', () => {
    hoisted.result = { data: undefined, isLoading: false, isError: true }
    render(<NetworkScreen />)
    expect(screen.getByText(/couldn.t load the network/i)).toBeInTheDocument()
  })

  it('renders the network once loaded', () => {
    hoisted.result = { data: facilities, isLoading: false, isError: false }
    render(<NetworkScreen />)
    expect(screen.getByRole('heading', { name: 'Network' })).toBeInTheDocument()
    expect(screen.getByText('Tafo Maternity')).toBeInTheDocument()
  })
})
