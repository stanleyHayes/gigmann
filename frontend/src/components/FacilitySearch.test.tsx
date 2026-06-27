import { render, screen, fireEvent } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { FacilityMatch } from '../api/useFacilitySearch'
import { FacilitySearch } from './FacilitySearch'

type SearchState = { data: FacilityMatch[]; isFetching: boolean }

const hoisted = vi.hoisted(() => ({
  result: { data: [] as FacilityMatch[], isFetching: false } as SearchState,
  navigate: vi.fn(),
}))

vi.mock('../api/useFacilitySearch', () => ({
  useFacilitySearch: () => hoisted.result,
}))
vi.mock('react-router-dom', async (importOriginal) => ({
  ...(await importOriginal<typeof import('react-router-dom')>()),
  useNavigate: () => hoisted.navigate,
}))

function openDialog() {
  render(<FacilitySearch />)
  fireEvent.click(screen.getByRole('button', { name: /search facilities/i }))
}

describe('FacilitySearch', () => {
  beforeEach(() => {
    hoisted.result = { data: [], isFetching: false }
    hoisted.navigate = vi.fn()
  })

  it('opens the dialog and prompts for input', () => {
    openDialog()
    expect(screen.getByText('Find a facility')).toBeInTheDocument()
    expect(screen.getByText(/type a name or a natural-language phrase/i)).toBeInTheDocument()
  })

  it('renders ranked matches and navigates on select', () => {
    hoisted.result = {
      data: [
        { facilityId: 'kasoa', name: 'Kasoa Polyclinic', score: 0.82 },
        { facilityId: 'nima', name: 'Nima Urban Health Centre', score: 0.4 },
      ],
      isFetching: false,
    }
    openDialog()
    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'kasoa' } })
    expect(screen.getByText('Kasoa Polyclinic')).toBeInTheDocument()
    expect(screen.getByText('82% match')).toBeInTheDocument()

    fireEvent.click(screen.getByText('Kasoa Polyclinic'))
    expect(hoisted.navigate).toHaveBeenCalledWith('/facilities/kasoa')
  })

  it('selects the first match on Enter', () => {
    hoisted.result = {
      data: [{ facilityId: 'kasoa', name: 'Kasoa Polyclinic', score: 0.82 }],
      isFetching: false,
    }
    openDialog()
    const field = screen.getByRole('textbox')
    fireEvent.change(field, { target: { value: 'kas' } })
    fireEvent.keyDown(field, { key: 'Enter' })
    expect(hoisted.navigate).toHaveBeenCalledWith('/facilities/kasoa')
  })

  it('shows a no-match message for an unmatched query', () => {
    hoisted.result = { data: [], isFetching: false }
    openDialog()
    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'zzz' } })
    expect(screen.getByText(/no facilities match/i)).toBeInTheDocument()
  })
})
