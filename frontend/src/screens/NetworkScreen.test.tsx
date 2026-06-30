import { fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
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
    render(<NetworkOverview facilities={facilities} />, { wrapper: MemoryRouter })
    expect(screen.getByText(/3 facilities/i)).toBeInTheDocument()
    expect(screen.getByText(/1 critical/i)).toBeInTheDocument()

    const names = screen.getAllByRole('heading', { level: 6 }).map((h) => h.textContent)
    expect(names[0]).toMatch(/Tafo/) // critical first
    expect(names[names.length - 1]).toMatch(/Asokwa/) // healthy last
  })

  it('handles an empty network', () => {
    render(<NetworkOverview facilities={[]} />, { wrapper: MemoryRouter })
    expect(screen.getByText(/0 facilities/i)).toBeInTheDocument()
  })

  it('paginates facility cards when the network grows', () => {
    const manyFacilities: Facility[] = Array.from({ length: 10 }, (_, i) => {
      const label = String(i + 1).padStart(2, '0')
      return {
        id: `facility-${label}`,
        name: `Facility ${label}`,
        region: 'Greater Accra',
        town: `Town ${label}`,
        beds: 20 + i,
        status: 'good',
      }
    })

    render(<NetworkOverview facilities={manyFacilities} />, { wrapper: MemoryRouter })

    expect(screen.getByText('Facility 01')).toBeInTheDocument()
    expect(screen.queryByText('Facility 10')).not.toBeInTheDocument()
    expect(screen.getByText(/1-9 of 10 facilities/i)).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /go to page 2/i }))

    expect(screen.getByText('Facility 10')).toBeInTheDocument()
    expect(screen.queryByText('Facility 01')).not.toBeInTheDocument()
  })
})

describe('NetworkScreen', () => {
  beforeEach(() => {
    hoisted.result = { data: undefined, isLoading: true, isError: false }
  })

  it('shows a skeleton while loading', () => {
    render(<NetworkScreen />, { wrapper: MemoryRouter })
    expect(screen.getByTestId('network-skeleton')).toBeInTheDocument()
  })

  it('shows an error state', () => {
    hoisted.result = { data: undefined, isLoading: false, isError: true }
    render(<NetworkScreen />, { wrapper: MemoryRouter })
    expect(screen.getByText(/couldn.t load the network/i)).toBeInTheDocument()
  })

  it('renders the network once loaded', () => {
    hoisted.result = { data: facilities, isLoading: false, isError: false }
    render(<NetworkScreen />, { wrapper: MemoryRouter })
    expect(screen.getByRole('heading', { name: 'Network' })).toBeInTheDocument()
    expect(screen.getByText('Tafo Maternity')).toBeInTheDocument()
  })
})
