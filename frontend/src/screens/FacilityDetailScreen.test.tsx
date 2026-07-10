import { fireEvent, render, screen, within } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { FacilityDetail } from '../api/useFacilityDetail'
import { FacilityDetailScreen } from './FacilityDetailScreen'

const hoisted = vi.hoisted(() => ({
  result: { data: undefined as FacilityDetail | undefined, isLoading: true, isError: false },
  taskMutate: vi.fn(),
  draftMutate: vi.fn(),
  draft: undefined as { draft: string } | undefined,
}))

vi.mock('../api/useFacilityDetail', () => ({
  useFacilityDetail: () => hoisted.result,
}))
vi.mock('../api/useTasks', () => ({
  useCreateTask: () => ({ mutate: hoisted.taskMutate, isPending: false, isError: false }),
}))
vi.mock('../api/useDrafts', () => ({
  useCreateDraft: () => ({ mutate: hoisted.draftMutate, isPending: false, isError: false, data: hoisted.draft }),
}))

function renderScreen() {
  const router = createMemoryRouter(
    [{ path: '/facilities/:facilityId', element: <FacilityDetailScreen /> }],
    { initialEntries: ['/facilities/kasoa'] },
  )
  return render(<RouterProvider router={router} />)
}

const detail: FacilityDetail = {
  facility: { id: 'kasoa', name: 'Kasoa Polyclinic', region: 'Central', town: 'Kasoa', beds: 40, status: 'critical' },
  kpis: [],
  inventory: [{ id: 'i1', name: 'RDT kits', category: 'consumable', stock_level: 20, days_of_stock: 3, stockout_imminent: true }],
  staff: [{ id: 's1', name: 'Ama Owusu', role: 'Medical Officer', status: 'active', attrition_risk: 0.7 }],
  alerts: [{ id: 'a1', type: 'denial_spike', severity: 'critical', title: 'NHIS denials high', status: 'open' }],
}

describe('FacilityDetailScreen', () => {
  beforeEach(() => {
    hoisted.result = { data: undefined, isLoading: true, isError: false }
    hoisted.taskMutate.mockReset()
    hoisted.draftMutate.mockReset()
    hoisted.draft = undefined
  })

  it('shows a skeleton while loading', () => {
    renderScreen()
    expect(screen.getByTestId('facility-skeleton')).toBeInTheDocument()
  })

  it('shows an error state', () => {
    hoisted.result = { data: undefined, isLoading: false, isError: true }
    renderScreen()
    expect(screen.getByText(/couldn.t load this facility/i)).toBeInTheDocument()
  })

  it('renders the facility and its sub-resources', () => {
    hoisted.result = { data: detail, isLoading: false, isError: false }
    renderScreen()
    expect(screen.getByRole('heading', { name: /Kasoa Polyclinic/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /add follow-up task/i })).toBeInTheDocument()
    expect(screen.getByText('RDT kits')).toBeInTheDocument()
    expect(screen.getByText(/Stockout imminent/i)).toBeInTheDocument()
    expect(screen.getByText(/NHIS denials high/i)).toBeInTheDocument()
    expect(screen.getByText(/Attrition risk/i)).toBeInTheDocument()
  })

  it('paginates growable facility detail sections', () => {
    hoisted.result = {
      data: {
        ...detail,
        staff: Array.from({ length: 6 }, (_, i) => ({
          id: `s-${i + 1}`,
          name: `Staff Member ${i + 1}`,
          role: 'Nurse',
          status: 'active',
          attrition_risk: 0.1,
        })),
      },
      isLoading: false,
      isError: false,
    }

    renderScreen()

    expect(screen.getByText(/Staff Member 1/i)).toBeInTheDocument()
    expect(screen.queryByText(/Staff Member 6/i)).not.toBeInTheDocument()

    fireEvent.click(within(screen.getByTestId('facility-staff-pagination')).getByRole('button', { name: /go to page 2/i }))

    expect(screen.getByText(/Staff Member 6/i)).toBeInTheDocument()
    expect(screen.queryByText(/Staff Member 1/i)).not.toBeInTheDocument()
  })

  it('wires the follow-up-task, draft, and copy actions', async () => {
    hoisted.result = { data: detail, isLoading: false, isError: false }
    hoisted.draft = { draft: 'Dear Ama, please prioritise the RDT reorder.' }
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true })

    renderScreen()

    fireEvent.click(screen.getByRole('button', { name: /add follow-up task/i }))
    expect(hoisted.taskMutate).toHaveBeenCalledTimes(1)

    fireEvent.click(screen.getByRole('button', { name: /draft manager message/i }))
    expect(hoisted.draftMutate).toHaveBeenCalledTimes(1)

    // The draft is present, so the Copy control renders and copies it.
    fireEvent.click(screen.getByRole('button', { name: /^copy$/i }))
    expect(writeText).toHaveBeenCalledWith('Dear Ama, please prioritise the RDT reorder.')
  })
})
